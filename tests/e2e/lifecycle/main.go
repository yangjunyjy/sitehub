// Lifecycle fixture exercises the same Cobra/go-zero/governance path as services
// without database dependencies. It is never included in deployment manifests.
package main

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/rest"
	"sitehub/common/command"
	"sitehub/common/governance"
	"net/http"
	"os"
	"time"
)

type config struct {
	governance.Config
	rest.RestConf
}

func main() { command.Execute(command.New("lifecycle-api", "lifecycle.yaml", run)) }
func run(parent context.Context, c config) error {
	c.Middlewares.Log = false
	s := rest.MustNewServer(c.RestConf)
	defer s.Stop()
	managed := governance.FromContext(parent)
	managed.AttachHTTP(s)
	s.AddRoute(rest.Route{Method: "GET", Path: "/test/pid", Handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, os.Getpid()) }})
	s.AddRoute(rest.Route{Method: "GET", Path: "/test/stream", Handler: func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		for i := 0; i < 12; i++ {
			if _, err := fmt.Fprintln(w, "chunk"); err != nil {
				return
			}
			_ = http.NewResponseController(w).Flush()
			select {
			case <-r.Context().Done():
				return
			case <-time.After(time.Second):
			}
		}
		_, _ = fmt.Fprintln(w, "complete")
	}}, rest.WithTimeout(20*time.Second))
	return managed.Serve(parent, fmt.Sprintf("%s:%d", c.Host, c.Port), s.Start)
}
