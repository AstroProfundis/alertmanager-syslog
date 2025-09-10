package webhook

import (
	"io"
	"net/http"

	"github.com/AstroProfundis/alertmanager-syslog/pkg/version"
	"github.com/golang/glog"
)

// ShowVersion returns the version of this program
func (s *Server) ShowVersion(w http.ResponseWriter, req *http.Request) {
	defer func() {
		err := req.Body.Close()
		if err != nil {
			glog.Errorf("Error closing request: %v", err)
		}
	}()

	metricRequestTotal.Inc()

	if req.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	if _, err := io.WriteString(w, version.NewVersion().String()); err != nil {
		glog.Errorf("Error sending version response: %v", err)
	}
}
