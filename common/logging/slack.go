package logging

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nlopes/slack"
)

const (
	textLimit   = 8000
	attachLimit = 100
	sendDelay   = time.Second
)

var (
	colorMap = map[Level]string{
		Critical: "danger",
		Error:    "danger",
		Warn:     "warning",
		Info:     "good",
		Debug:    "#eeeeee",
	}

	// Singleton for each channel.
	// Key is channel ID. Value is SlackOutput.
	buffers    sync.Map
	workerOnce sync.Once
)

// SlackOpt returns a new slack opt.
func SlackOpt() *SlackOption {
	return &SlackOption{}
}

// SlackOption defines the slack option struct.
type SlackOption struct {
	code bool
}

// Code sets if the log is output in code block.
func (o *SlackOption) Code(c bool) *SlackOption {
	o.code = c
	return o
}

type outFunc func() Output

// NewSlackOutput returns a slack output.
func NewSlackOutput(ctx context.Context, api *slack.Client, channel string) Output {
	if out, ok := buffers.Load(channel); ok {
		return (out.(outFunc))()
	}

	// A safe way to initialize.
	var out *slackOutput
	var once sync.Once

	ofn := outFunc(func() Output {
		once.Do(
			// Initializing.
			func() {
				out = &slackOutput{
					ctx:        ctx,
					api:        api,
					channel:    channel,
					defaultOpt: SlackOpt(),
				}
			})
		return out
	})

	fn, loaded := buffers.LoadOrStore(channel, ofn)
	if loaded {
		f := fn.(outFunc)
		return f()
	}

	workerOnce.Do(func() {
		go work()
	})

	return ofn()
}

type slackOutput struct {
	ctx     context.Context
	api     *slack.Client
	channel string

	mtx     sync.Mutex
	outBuff []*slack.Attachment

	defaultOpt *SlackOption
}

func (o *slackOutput) Output(
	opt *OutputOpt, level Level, labelMap LabelMap, log string) {
	slackOpt := o.defaultOpt
	if slackOptIn, ok := (*sync.Map)(opt).Load("slack"); ok {
		slackOpt = slackOptIn.(*SlackOption)
	}
	a := &slack.Attachment{
		Text:  removeColor(log),
		Color: colorMap[level],
	}
	if slackOpt.code {
		a.Text = fmt.Sprintf("```\n%s```", a.Text)
	}
	if tag, ok := labelMap[LabelTag]; ok {
		if !slackOpt.code {
			a.Text = fmt.Sprintf("*%s* | %s", tag, a.Text)
			a.MarkdownIn = []string{"text"}
		} else {
			a.Title = tag
		}
	}
	if level <= Error {
		a.Footer = labelMap.debugInfo(false)
	}

	o.mtx.Lock()
	o.outBuff = append(o.outBuff, a)
	o.mtx.Unlock()
}

func work() {
	var landing bool
	var chanCount int
	for !(landing && chanCount == 0) {
		chanCount = 0
		// All channels shares one global lock to satisfy with rate limit.
		buffers.Range(func(k, v interface{}) bool {
			chanCount++

			fn := v.(outFunc)
			out := fn()
			o := out.(*slackOutput)

			// If we sent anything, sleep due to rate limit.
			if o.flush() != 0 {
				time.Sleep(sendDelay)
			} else {
				// It becomes empty.
				chanCount--
			}

			// If any ctx is Done, starts landing mode.
			select {
			case <-o.ctx.Done():
				buffers.Delete(k)
				landing = true
			default:
			}

			return true
		})
	}
}

func (o *slackOutput) flush() (ret int) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if o.outBuff == nil {
		return
	}

	var charCount int
	msgBuff := make([]slack.Attachment, 0, 5)
	for i := 0; i < len(o.outBuff) &&
		charCount < textLimit &&
		len(msgBuff) < attachLimit; i++ {
		ptr := o.outBuff[i]

		charCount += len(ptr.Text)
		// If color is equals to the last msgs, merge them.
		if len(msgBuff) != 0 &&
			msgBuff[len(msgBuff)-1].Color == ptr.Color {
			msgBuff[len(msgBuff)-1].Text = strings.Replace(
				fmt.Sprintf("%s\n%s", msgBuff[len(msgBuff)-1].Text, ptr.Text),
				"\n```\n```", "", -1)
		} else {
			msgBuff = append(msgBuff, *ptr)
		}
	}

	ret = len(msgBuff)
	o.outBuff = o.outBuff[:0]
	if len(msgBuff) > 0 {
		o.sendMulti(msgBuff)
	}

	return
}

func (o *slackOutput) sendMulti(as []slack.Attachment) bool {
	_, _, _, err := o.api.SendMessageContext(
		o.ctx,
		o.channel,
		slack.MsgOptionAttachments(as...),
	)
	if err != nil {
		fmt.Println("failed to post msg to slack", o.channel, err)
		return false
	}

	return true
}
