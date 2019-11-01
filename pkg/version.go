package webhook

import (
	"fmt"
	"net/http"

	"github.com/AstroProfundis/alertmanager-syslog/pkg/version"
)

// ShowVersion returns the version of this program
func (s *Server) ShowVersion(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	metricRequestTotal.Inc()

	if req.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, version.NewVersion().String())
}
