package logging

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/cobinhood/mochi/common/config/misc"
)

type stackdriverOutput struct {
	client *logging.Client
	logger *logging.Logger
}

func newStackdriverOutput(logname string) (*stackdriverOutput, error) {
	ctx := context.Background()
	client, err := logging.NewClient(
		ctx, misc.ServerProjectId())
	if err != nil {
		return nil, err
	}
	// Check if Connection is Valid.
	err = client.Ping(ctx)
	if err != nil {
		return nil, err
	}
	o := &stackdriverOutput{client: client}
	o.refreshLogger(logname)
	return o, nil
}

func (o *stackdriverOutput) refreshLogger(logname string) {
	if o.logger != nil {
		return
	}
	if logname == "" {
		return
	}
	// logName="projects/cobinhood/logs/${logname}" and
	// resource.type="gce_instance" on StackDriver
	o.logger = o.client.Logger(logname)
}

func (o *stackdriverOutput) Output(
	opt *OutputOpt, level Level, labelMap LabelMap, log string) {
	if o.logger == nil {
		return
	}
	o.logger.Log(logging.Entry{
		Severity: level.Severity(),
		Labels:   labelMap,
		Payload:  removeColor(log),
	})
}
