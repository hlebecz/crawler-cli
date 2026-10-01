package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	page := func(title, link string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			body := fmt.Sprintf(`<html><head><title>%s</title></head><body>`, title)
			if link != "" {
				body += fmt.Sprintf(`<a href="%s">next</a>`, link)
			}
			body += `</body></html>`
			fmt.Fprint(w, body)
		}
	}

	mux.Handle("/", page("Root", "/child"))
	mux.Handle("/child", page("Child", "/grandchild"))
	mux.Handle("/grandchild", page("Grandchild", ""))

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
