package profiler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	unknownRoutes := []string{
		"foo", "bar",
	}
	routes := []string{
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
	var called []string
	h := New(Options{
		Routes: append(routes, unknownRoutes...),
		UnknownRoute: func(route string) {
			called = append(called, route)
		},
	})
	require.NotNil(t, h)
	require.Equal(t, unknownRoutes, called)
}

func TestHandler_ServeHTTP(t *testing.T) {
	h := New(Options{})
	require.NotNil(t, h)
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)

	req, err := http.NewRequest(http.MethodGet, s.URL+"/debug/pprof/", nil)
	require.NoError(t, err)

	res, err := s.Client().Do(req)
	require.NoErrorf(t, err, "request: %s", req.URL)
	require.Equalf(t, http.StatusOK, res.StatusCode, "status: %s", res.Status)
}
