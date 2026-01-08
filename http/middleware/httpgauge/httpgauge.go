package httpgauge

import (
	"github.com/go-chi/chi/v5"
	"maps"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Gauge struct {
	stats map[string]int
	mu    sync.Mutex
}

func New() *Gauge {
	return &Gauge{
		stats: make(map[string]int),
	}
}

func (g *Gauge) Snapshot() map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	snapshot := make(map[string]int, len(g.stats))
	maps.Copy(snapshot, g.stats)
	return snapshot
}

func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snapshot := g.Snapshot()

	keys := make([]string, 0, len(snapshot))
	for k := range snapshot {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString(" ")
		builder.WriteString(strconv.Itoa(snapshot[k]))
		builder.WriteString("\n")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(builder.String()))
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pattern := getRoutePattern(r)
			g.mu.Lock()
			g.stats[pattern]++
			g.mu.Unlock()
		}()

		next.ServeHTTP(w, r)
	})
}

func getRoutePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}
