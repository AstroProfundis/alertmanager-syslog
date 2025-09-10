package webhook

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	json "github.com/json-iterator/go"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricAlertProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alertmanager_syslog_alert_total",
		Help: "The total number of processed alerts",
	})
	metricAlertSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alertmanager_syslog_alert_sent",
		Help: "The number of alerts that sent to Syslog server",
	})
	metricAlertSentError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alertmanager_syslog_alert_sent_error",
		Help: "The number of alerts that failed sending to Syslog server",
	})
	metricRequestTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alertmanager_syslog_requests_total",
		Help: "The total number of received requests",
	})
	metricRequestError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alertmanager_syslog_requests_error",
		Help: "The number of received requests that are unable to process",
	})
)

// HandleAlert handles webhook for AlertManager
func (s *Server) HandleAlert(w http.ResponseWriter, req *http.Request) {
	defer func() {
		err := req.Body.Close()
		if err != nil {
			glog.Errorf("Error closing request: %v", err)
		}
	}()

	metricRequestTotal.Inc()

	data := template.Data{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		errMsg := fmt.Sprintf("Error parsing request: %s", err)
		http.Error(w, errMsg, http.StatusBadRequest)
		glog.Error(errMsg)
		metricRequestError.Inc()
		return
	}

	if err := s.sendAlert(data); err != nil {
		errMsg := fmt.Sprintf("Error sending message to syslog server: %v", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		glog.Error(errMsg)
		metricRequestError.Inc()
	}
}

func (s *Server) sendAlert(data template.Data) error {
	// check for mode
	if s.config.Mode == "" {
		return fmt.Errorf("mode is not set")
	}

	commLabels := strings.Join(data.CommonLabels.Values(), "|")
	for _, alert := range data.Alerts {
		metricAlertProcessed.Inc()

		var msg []byte
		var err error
		// parse and build message for custom mode
		if strings.ToLower(s.config.Mode) == "custom" {
			msg, err = s.customMsg(alert)
		} else {
			// build message for default modes
			msg, err = s.sysLogMsg(alert, commLabels)
		}
		if err != nil {
			metricAlertSentError.Inc()
			return err
		}

		if _, err = s.sysLog.Write(msg); err != nil {
			metricAlertSentError.Inc()
			return err
		}

		metricAlertSent.Inc()
		glog.V(3).Infof("Send alert: [%s] %s\n", getAlertValue(alert.Labels, "severity"), msg)
	}
	return nil
}
