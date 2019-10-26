package webhook

import (
	"context"
	"fmt"
	"log/syslog"
	"net/http"
	"time"
)

// ServerCfg is the config options used by Server
type ServerCfg struct {
	ListenAddr  string
	SyslogAddr  string
	Network     string
	Timeout     int
	Mode        string
	Labels      []string
	Annotations []string
}

// Server is a webhook server to handle messages
type Server struct {
	httpServer *http.Server
	sysLog     *syslog.Writer

	mode        string
	labels      []string
	annotations []string
}

// New create a Server
func New(cfg *ServerCfg) (*Server, error) {
	timeoutSec := time.Second * time.Duration(cfg.Timeout)

	syslogWriter, err := syslog.Dial(cfg.Network,
		cfg.SyslogAddr,
		syslog.LOG_CRIT|syslog.LOG_USER,
		"alertmanager-syslog")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Listening on %s, timeout is %v\n", cfg.ListenAddr, timeoutSec)
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ListenAddr,
			ReadTimeout:  timeoutSec,
			WriteTimeout: timeoutSec,
		},
		sysLog:      syslogWriter,
		mode:        cfg.Mode,
		labels:      cfg.Labels,
		annotations: cfg.Annotations,
	}, nil
}

// ListenAndServe starts the server
func (s *Server) ListenAndServe() {
	s.httpServer.ListenAndServe()
}

// Close closes the server
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		cancel()
	}()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		fmt.Println("Failed to shutdown HTTP server, closing anyway.")
		s.httpServer.Close() // ignore the error
	} else {
		fmt.Println("Finish shuting down HTTP server.")
	}

	s.sysLog.Close() // ignore the error
	fmt.Println("Closed connection to Syslog server.")
}
