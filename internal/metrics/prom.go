package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Registry struct {
	events uint64
	errors uint64
}

func (r *Registry) IncEvents() { atomic.AddUint64(&r.events, 1) }
func (r *Registry) IncErrors() { atomic.AddUint64(&r.errors, 1) }

func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "sequids_events_total %d\nsequids_errors_total %d\n", atomic.LoadUint64(&r.events), atomic.LoadUint64(&r.errors))
	}
}
