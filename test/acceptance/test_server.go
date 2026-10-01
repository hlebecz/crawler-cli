package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serverFromMux(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func testMux() *http.ServeMux {
	mux := http.NewServeMux()
	page := func(title string, links []string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			var body strings.Builder
			body.WriteString(fmt.Sprintf(`<html><head><title>%s</title></head><body>`, title))
			if len(links) > 0 {
				for _, link := range links {
					body.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, link, link))
				}
			}
			body.WriteString(`</body></html>`)
			fmt.Fprint(w, body.String())
		}
	}

	mux.Handle("/", page("Root", []string{"/child"}))
	mux.Handle("/child", page("Child", []string{"/grandchild1", "/grandchild2"}))
	mux.Handle("/grandchild1", page("Grandchild1", []string{}))
	mux.Handle("/grandchild2", page("Grandchild2", []string{}))

	return mux
}

func slowMux(onRequest func(), delay time.Duration) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		onRequest()
		select {
		case <-r.Context().Done():
		case <-time.After(delay):
		}
	})
	return mux
}
