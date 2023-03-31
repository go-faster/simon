package app

import (
	"net/http/pprof"
	"os"
	"path"
	rpprof "runtime/pprof"
	"strings"

	"go.uber.org/zap"
)

func (m *Metrics) registerProfiler() {
	var routes []string
	if v := os.Getenv("PPROF_ROUTES"); v != "" {
		routes = strings.Split(v, ",")
	}
	if len(routes) == 1 && routes[0] == "none" {
		return
	}
	if len(routes) == 0 {
		// Enable all routes by default except cmdline (unsafe).
		//
		// Route name is "/debug/pprof/<name>".
		routes = []string{
			// From pprof.<Name>.
			"profile",
			"symbol",
			"trace",

			// From pprof.Handler(<name>).
			"goroutine",
			"heap",
			"threadcreate",
			"block",
		}
	}
	m.lg.Info("Registering pprof routes", zap.Strings("routes", routes))
	m.mux.HandleFunc("/debug/pprof/", pprof.Index)
	for _, name := range routes {
		name = strings.TrimSpace(name)
		route := path.Join("/debug/pprof/", name)
		switch name {
		case "cmdline":
			m.mux.HandleFunc(route, pprof.Cmdline)
		case "profile":
			m.mux.HandleFunc(route, pprof.Profile)
		case "symbol":
			m.mux.HandleFunc(route, pprof.Symbol)
		case "trace":
			m.mux.HandleFunc(route, pprof.Trace)
		case "none": // invalid
			m.lg.Warn("Invalid pprof route ('none' should be the only one route specified)",
				zap.String("route", name),
			)
		default:
			if rpprof.Lookup(name) == nil {
				m.lg.Warn("Invalid pprof route", zap.String("route", name))
				continue
			}
			m.mux.Handle(route, pprof.Handler(name))
		}
	}
}
