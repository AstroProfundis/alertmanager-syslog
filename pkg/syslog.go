package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	syslog "github.com/RackSec/srslog"
	"github.com/golang/glog"
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
	var severity string
	switch strings.ToLower(s.config.Custom.Severities.Type) {
	case "label":
		severity = strings.ToUpper(getAlertValue(alert.Labels,
			s.config.Custom.Severities.Key))
	case "annotation":
		severity = strings.ToUpper(getAlertValue(alert.Annotations,
			s.config.Custom.Severities.Key))
	}

	valueList := make([]string, 0)
	for _, sect := range s.config.Custom.Sections {
		var columns []column
		if sect.Join {
			columns = sect.Columns
		} else {
			columns = sect.Columns[0:1:1] // only get the first column
		}

		// parse columns
		colValues := make([]string, 0)
		for _, col := range columns {
			var colValRaw string
			switch strings.ToLower(col.Type) {
			case "const":
				colValRaw = col.Value
			case "label":
				colValRaw = getAlertValue(alert.Labels, col.Key)
			case "annotation":
				colValRaw = getAlertValue(alert.Annotations, col.Key)
			case "time":
				if alert.Status == "resolved" {
					colValRaw = strconv.FormatInt(alert.EndsAt.Unix(), 10)
				} else {
					colValRaw = strconv.FormatInt(alert.StartsAt.Unix(), 10)
				}
			case "instance":
				instance := getAlertValue(alert.Labels, "instance")
				if col.StripPort {
					instance = strings.Split(instance, ":")[0]
				}
				colValRaw = instance
			case "status":
				colValRaw = alert.Status
			case "severity":
				// treat resolved status as a special severity
				if s.config.Custom.Severities.IncludeResolved &&
					alert.Status == "resolved" {
					severity = alert.Status
				}
				colValRaw = parseSeverity(severity, &s.config.Custom.Severities)
				// try to parse empty severity again if replace empty is set, as
				// it might be defined as a special numeric severity
				if s.config.Custom.ReplaceEmpty != "" && colValRaw == "" {
					colValRaw = parseSeverity(s.config.Custom.ReplaceEmpty, &s.config.Custom.Severities)
				}
			default:
				return nil, fmt.Errorf("Unknown section type")
			}

			// replace empty values to user defined placeholder
			if s.config.Custom.ReplaceEmpty != "" && len(colValRaw) < 1 {
				colValRaw = s.config.Custom.ReplaceEmpty
			}
			colValues = append(colValues, colValRaw)
		}

		// join columns if needed
		var columnString string
		if sect.Join {
			delimiter := "_"
			if sect.Delimiter != "" {
				delimiter = sect.Delimiter
			}
			columnString = strings.Join(colValues, delimiter)
		} else {
			columnString = colValues[0]
		}

		// replace white spaces if needed
		if s.config.Custom.ReplaceWhitespace != "" {
			columnString = strings.Join(strings.Fields(columnString), s.config.Custom.ReplaceWhitespace)
		}
		valueList = append(valueList, columnString)
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

func parseSeverity(s string, cfg *severities) string {
	if cfg.Mode != "number" {
		return s
	}

	for _, lv := range cfg.Levels {
		if strings.ToUpper(s) == strings.ToUpper(lv.Name) {
			return strconv.Itoa(lv.Value)
		}
	}
	// treat severity parsing errors as empty results
	if cfg.ErrorAsEmpty {
		return ""
	}
	return "-1"
}

// Priority converts priority settings in config to syslog.Priority
func Priority(s string) (syslog.Priority, error) {
	switch strings.ToUpper(s) {
	// severity
	case "EMERG":
		return syslog.LOG_EMERG, nil
	case "ALERT":
		return syslog.LOG_ALERT, nil
	case "CRIT":
		return syslog.LOG_CRIT, nil
	case "ERR":
		return syslog.LOG_ERR, nil
	case "WARNING":
		return syslog.LOG_WARNING, nil
	case "NOTICE":
		return syslog.LOG_NOTICE, nil
	case "INFO":
		return syslog.LOG_INFO, nil
	case "DEBUG":
		return syslog.LOG_DEBUG, nil
	// facility
	case "KERN":
		return syslog.LOG_KERN, nil
	case "USER":
		return syslog.LOG_USER, nil
	case "MAIL":
		return syslog.LOG_MAIL, nil
	case "DAEMON":
		return syslog.LOG_DAEMON, nil
	case "AUTH":
		return syslog.LOG_AUTH, nil
	case "SYSLOG":
		return syslog.LOG_SYSLOG, nil
	case "LPR":
		return syslog.LOG_LPR, nil
	case "NEWS":
		return syslog.LOG_NEWS, nil
	case "UUCP":
		return syslog.LOG_UUCP, nil
	case "CRON":
		return syslog.LOG_CRON, nil
	case "AUTHPRIV":
		return syslog.LOG_AUTHPRIV, nil
	case "FTP":
		return syslog.LOG_FTP, nil
	case "LOCAL0":
		return syslog.LOG_LOCAL0, nil
	case "LOCAL1":
		return syslog.LOG_LOCAL1, nil
	case "LOCAL2":
		return syslog.LOG_LOCAL2, nil
	case "LOCAL3":
		return syslog.LOG_LOCAL3, nil
	case "LOCAL4":
		return syslog.LOG_LOCAL4, nil
	case "LOCAL5":
		return syslog.LOG_LOCAL5, nil
	case "LOCAL6":
		return syslog.LOG_LOCAL6, nil
	case "LOCAL7":
		return syslog.LOG_LOCAL7, nil
	default:
		msg := fmt.Sprintf("Unknown priority %s", s)
		glog.Error(msg)
		return 0, fmt.Errorf(msg)
	}
}
