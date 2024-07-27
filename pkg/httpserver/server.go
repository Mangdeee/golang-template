package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	DEFAULT_READ_TIMEOUT     = 75 * time.Second
	DEFAULT_WRITE_TIMEOUT    = 75 * time.Second
	DEFAULT_ADDRESS          = ":80"
	DEFAULT_SHUTDOWN_TIMEOUT = 3 * time.Second
)

type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

func NewServer(handler http.Handler, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  DEFAULT_READ_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
		Addr:         DEFAULT_ADDRESS,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: DEFAULT_SHUTDOWN_TIMEOUT,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
