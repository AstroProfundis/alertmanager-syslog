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
	ListenAddr string
	SyslogAddr string
	Network    string
	Tag        string
	Timeout    int
	Hostname   string
	Config     *config
}

// Server is a webhook server to handle messages
type Server struct {
	httpServer *http.Server
	sysLog     *syslog.Writer

	hostname string
	config   *config
}

// New create a Server
func New(cfg *ServerCfg) (*Server, error) {
	timeoutSec := time.Second * time.Duration(cfg.Timeout)

	glog.V(3).Infof("Read priority setting from config: %s, %s",
		cfg.Config.Severity, cfg.Config.Facility)
	syslogSeverity, err := Priority(cfg.Config.Severity)
	if err != nil {
		glog.Warning("Error parsing severity from config, using LOG_CRIT as severity for syslog")
		syslogSeverity = syslog.LOG_CRIT
	}
	syslogFacility, err := Priority(cfg.Config.Facility)
	if err != nil {
		glog.Warning("Error parsing facility from config, using LOG_USER as severity for syslog")
		syslogFacility = syslog.LOG_USER
	}
	glog.V(3).Infof("Using priority %d", syslogSeverity|syslogFacility)

	syslogWriter, err := syslog.Dial(cfg.Network,
		cfg.SyslogAddr,
		syslogSeverity|syslogFacility,
		cfg.Tag)
	if err != nil {
		return nil, err
	}
	syslogWriter.SetFormatter(syslog.RFC3164Formatter)
	if cfg.Hostname != "" {
		syslogWriter.SetHostname(cfg.Hostname)
	}
	glog.V(3).Infof("Connected to syslog server %s://%s",
		cfg.Network, cfg.SyslogAddr)

	glog.Infof("Listening on %s, timeout is %v\n", cfg.ListenAddr, timeoutSec)
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ListenAddr,
			ReadTimeout:  timeoutSec,
			WriteTimeout: timeoutSec,
		},
		sysLog: syslogWriter,
		config: cfg.Config,
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
