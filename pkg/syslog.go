package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

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

// sysLogMsg build a syslog message from alert for default output format
func (s *Server) sysLogMsg(alert template.Alert, commLabels string) ([]byte, error) {
	// msg is the message send to syslog server
	msg := make(map[string]string)

	// add default labels
	for _, label := range defaultLabels {
		msg[label] = getAlertValue(alert.Labels, label)
	}
	msg["status"] = alert.Status
	msg["time"] = alert.StartsAt.Format(timeFormat)

	// add labels and annotations from configuration
	for _, label := range s.config.Labels {
		msg[label] = getAlertValue(alert.Labels, label)
	}
	for _, annon := range s.config.Annotations {
		msg[annon] = getAlertValue(alert.Annotations, annon)
	}

	// add all common labels
	msg["commonLabels"] = commLabels

	switch strings.ToLower(s.config.Mode) {
	// convert to plain text format
	case "plain", "text":
		return formatPlain(msg), nil
	// default format is JSON
	case "json":
		fallthrough
	default:
		return json.Marshal(msg)
	}
}

func (s *Server) customMsg(alert template.Alert) ([]byte, error) {
	valueList := make([]string, 0)
	for _, sect := range s.config.Custom.Sections {
		switch strings.ToLower(sect.Type) {
		case "const":
			valueList = append(valueList, sect.Value)
		case "label":
			valueList = append(valueList, getAlertValue(alert.Labels, sect.Key))
		case "annotation":
			valueList = append(valueList, getAlertValue(alert.Annotations, sect.Key))
		case "time":
			valueList = append(valueList, strconv.FormatInt(alert.StartsAt.Unix(), 10))
		case "status":
			valueList = append(valueList, alert.Status)
		default:
			return nil, fmt.Errorf("Unknown section type")
		}
	}

	return []byte(strings.Join(valueList, s.config.Custom.Delimiter)), nil
}

func getAlertValue(kv template.KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}

func formatPlain(kv map[string]string) []byte {
	// sort the kv pairs with keys to make the output constant, note that
	// in JSON ourput format, the keys are automatically sorted
	var keys []string
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := new(bytes.Buffer)
	for _, k := range keys {
		fmt.Fprintf(b, "%s=%v ", k, kv[k])
	}
	return b.Bytes()
}
