package webhook

import (
	"context"
	"net/http"
	"time"

	syslog "github.com/RackSec/srslog"
	"github.com/golang/glog"
)

// ServerCfg is the config options used by Server
type ServerCfg struct {
	ListenAddr  string
	SyslogAddr  string
	Network     string
	Timeout     int
	Mode        string
	Hostname    string
	Labels      []string
	Annotations []string
}

// Server is a webhook server to handle messages
type Server struct {
	httpServer *http.Server
	sysLog     *syslog.Writer

	mode        string
	hostname    string
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
	syslogWriter.SetFormatter(syslog.RFC3164Formatter)
	if cfg.Hostname != "" {
		syslogWriter.SetHostname(cfg.Hostname)
	}

	glog.Infof("Listening on %s, timeout is %v\n", cfg.ListenAddr, timeoutSec)
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
		glog.Warningf("Failed to shutdown HTTP server, closing anyway.")
		s.httpServer.Close() // ignore the error
	} else {
		glog.Infof("Finish shuting down HTTP server.")
	}

	s.sysLog.Close() // ignore the error
	glog.Infof("Closed connection to Syslog server.")
}
