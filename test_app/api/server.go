package api

import (
    "expvar"
    "flag"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "sync"
    "time"

	"github.com/jiarung/mochi/test_app/database"
)

// Command-line flags.
var (
    httpAddr   = flag.String("http", ":8080", "Listen address")
    pollPeriod = flag.Duration("poll", 5*time.Second, "Poll period")
)

const baseChangeURL = "https://localhost/+/"

//func main() {
//    flag.Parse()
//    changeURL := fmt.Sprintf("%s", baseChangeURL)
//    http.Handle("/", NewServer(changeURL, *pollPeriod))
//    log.Fatal(http.ListenAndServe(*httpAddr, nil))
//}

func Run() {

    flag.Parse()
    changeURL := fmt.Sprintf("%s", baseChangeURL)
    http.Handle("/", NewServer(changeURL, *pollPeriod))
    log.Fatal(http.ListenAndServe(*httpAddr, nil))
    database.Initialize("default")

    seed.Load(server.DB)

    server.Run(":8080")
}

// Exported variables for monitoring the server.
// These are exported via HTTP as a JSON object at /debug/vars.
var (
    hitCount       = expvar.NewInt("hitCount")
    pollCount      = expvar.NewInt("pollCount")
    pollError      = expvar.NewString("pollError")
    pollErrorCount = expvar.NewInt("pollErrorCount")
)

// Server implements the outyet server.
// It serves the user interface (it's an http.Handler)
// and polls the remote repository for changes.
type Server struct {
    url     string
    period  time.Duration

    mu  sync.RWMutex // protects the success variable
    success bool
}

// NewServer returns an initialized outyet server.
func NewServer(url string, period time.Duration) *Server {
    s := &Server{url: url, period: period}
    go s.poll()
    return s
}

// poll polls the change URL for the specified period until the tag exists.
// Then it sets the Server's success field true and exits.
func (s *Server) poll() {
    for !isTagged(s.url) {
        pollSleep(s.period)
    }
    s.mu.Lock()
    s.success = true
    s.mu.Unlock()
    pollDone()
}

// Hooks that may be overridden for integration tests.
var (
    pollSleep = time.Sleep
    pollDone  = func() {}
)

// isTagged makes an HTTP HEAD request to the given URL and reports whether it
// returned a 200 OK response.
func isTagged(url string) bool {
    pollCount.Add(1)
    r, err := http.Head(url)
    if err != nil {
        log.Print(err)
        pollError.Set(err.Error())
        pollErrorCount.Add(1)
        return false
    }
    return r.StatusCode == http.StatusOK
}

// ServeHTTP implements the HTTP user interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    hitCount.Add(1)
    s.mu.RLock()
    data := struct {
        URL     string
        Success     bool
    }{
        s.url,
        s.success,
    }
    fmt.Printf("%s\n", *httpAddr)
    s.mu.RUnlock()
    err := tmpl.Execute(w, data)
    if err != nil {
        log.Print(err)
    }
}

// tmpl is the HTML template that drives the user interface.
var tmpl = template.Must(template.New("tmpl").Parse(`
<!DOCTYPE html><html><body><center>
    <h2> hello dcard! </h2>
    <h1>
    {{if .Success}}
        <a href="{{.URL}}">Ping!</a>
    {{else}}
        Rate limited :-(
    {{end}}
    </h1>
</center></body></html>
`))
