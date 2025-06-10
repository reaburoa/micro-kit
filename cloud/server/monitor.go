package server

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/felixge/fgprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/welltop-cn/common/utils/log"
)

func RunMetrics() error {
	metricsHTTPPort := os.Getenv("METRICS_HTTP_PORT")
	if metricsHTTPPort == "" {
		metricsHTTPPort = ":3000"
	}

	go func() {
		metricsSrv := http.NewServeMux()
		metricsSrv.Handle("/metrics", promhttp.Handler())
		metricsSrv.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("pong"))
		}))
		if err := http.ListenAndServe(metricsHTTPPort, metricsSrv); err != nil {
			log.Fatalf("start monitor http server failed %+v", err)
		}
	}()

	return nil
}

func RunPprof() error {
	pprofHTTPPort := os.Getenv("PPROF_HTTP_PORT")
	if pprofHTTPPort == "" {
		pprofHTTPPort = ":9090"
	}
	go func() {
		pprofSrv := http.NewServeMux()
		pprofSrv.HandleFunc("/debug/pprof/", pprof.Index)
		pprofSrv.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofSrv.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofSrv.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofSrv.HandleFunc("/debug/pprof/trace", pprof.Trace)
		if err := http.ListenAndServe(pprofHTTPPort, pprofSrv); err != nil {
			log.Fatalf("start pprof http server failed %+v", err)
		}
	}()
	fgprofHTTPPort := os.Getenv("FGPROF_HTTP_PORT")
	if fgprofHTTPPort == "" {
		fgprofHTTPPort = ":6060"
	}
	go func() {
		fgprofSrv := http.NewServeMux()
		fgprofSrv.Handle("/debug/fgprof", fgprof.Handler())
		if err := http.ListenAndServe(fgprofHTTPPort, fgprofSrv); err != nil {
			log.Fatalf("start fgprof http server failed %+v", err)
		}
	}()

	return nil
}
