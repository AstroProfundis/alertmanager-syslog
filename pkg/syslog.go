package webhook

import (
	"encoding/json"

	"github.com/prometheus/alertmanager/template"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

// syslogMessage syslog base struct
type syslogMessage struct {
	AlertName string `json:"alertname"`
	Instance  string `json:"instance"`
	Level     string `json:"level"`
	Status    string `json:"status"`
	Time      string `json:"time"`
	Labels    string `json:"labels"`
}

// sysLogMsg build a syslog message from alert
func sysLogMsg(alert template.Alert, labels string) ([]byte, error) {
	msg := &syslogMessage{
		AlertName: getValue(alert.Labels, "alertname"),
		Level:     getValue(alert.Labels, "level"),
		Instance:  getValue(alert.Annotations, "instance"),
		Status:    alert.Status,
		Time:      alert.StartsAt.Format(timeFormat),
		Labels:    labels,
	}
	return json.Marshal(*msg)
}

func getValue(kv template.KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}
