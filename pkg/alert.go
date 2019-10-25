package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/alertmanager/template"
)

// HandleAlert handles webhook for AlertManager
func (s *Server) HandleAlert(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	data := template.Data{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %s", err), http.StatusBadRequest)
		return
	}

	go func() {
		if err := s.sendAlert(data); err != nil {
			http.Error(w,
				fmt.Sprintf("Error sending message to syslog server: %v", err),
				http.StatusInternalServerError)
		}
	}()
}

func (s *Server) sendAlert(data template.Data) error {
	commLabels := strings.Join(data.CommonLabels.Values(), "|")
	for _, alert := range data.Alerts {
		severity := strings.ToUpper(getValue(alert.Labels, "severity"))
		switch severity {
		case "CRITICAL", "WARNING":
			msg, err := s.sysLogMsg(alert, commLabels)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(s.sysLog, "%s", msg)
			if err != nil {
				return err
			}

			fmt.Printf("[%s] %s\n", getValue(alert.Labels, "severity"), msg)
		}
	}
	return nil
}
