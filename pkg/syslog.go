package webhook

import (
	"encoding/json"

	"github.com/prometheus/alertmanager/template"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

var (
	// defaultLabels are labels always been added to the message
	defaultLabels = [...]string{
		"alertname",
		"severity",
	}
)

// sysLogMsg build a syslog message from alert
func (s *Server) sysLogMsg(alert template.Alert, commLabels string) ([]byte, error) {
	// msg is the message send to syslog server
	msg := make(map[string]string)

	// add default labels
	for _, label := range defaultLabels {
		msg[label] = getValue(alert.Labels, label)
	}
	msg["status"] = alert.Status
	msg["time"] = alert.StartsAt.Format(timeFormat)

	// add labels and annotations from configuration
	for _, label := range s.labels {
		msg[label] = getValue(alert.Labels, label)
	}
	for _, annon := range s.annotations {
		msg[annon] = getValue(alert.Annotations, annon)
	}

	// add all common labels
	msg["commonLabels"] = commLabels

	return json.Marshal(msg)
}

func getValue(kv template.KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}
