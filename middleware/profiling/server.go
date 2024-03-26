package profiling

import (
	"context"
	"net"
	"net/http"
	"os"
	"runtime/trace"
	"time"

	"github.com/rs/zerolog"
)

// Serve starts the pprof server
// It should run in a goroutine du to the blocking nature of the server
func Serve(log *zerolog.Logger) {
	log.Info().Msg("Starting pprof server")
	// add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	handler := http.DefaultServeMux

	http.DefaultServeMux = http.NewServeMux()

	http.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

	srv := &http.Server{
		Addr:              "localhost:6060",
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           handler,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to start pprof server")
	}
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := trace.Start(f); err != nil {
		log.Panic().Err(err).Msg("Failed to start trace")
	}
	trace.Stop()
}
