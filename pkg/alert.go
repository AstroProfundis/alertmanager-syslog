package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/alertmanager/template"
)

type errResp struct {
	Status  string
	Message string
}

// HandleAlert handles webhook for AlertManager
func HandleAlert(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	data := template.Data{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %s", err), http.StatusBadRequest)
		return
	}
	fmt.Printf("%v\n", data)

	for _, alert := range data.Alerts.Firing() {
		severity := alert.Labels["severity"]
		switch strings.ToUpper(severity) {
		case "CRITICAL", "WARNING":
			sendAlert(alert)
		}
	}
}

func sendAlert(alert template.Alert) {
	fmt.Printf("%v\n", alert)
}
