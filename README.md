# AlertManager Syslog Receiver

[![Build Status](https://travis-ci.org/AstroProfundis/alertmanager-syslog.svg?branch=master)](https://travis-ci.org/AstroProfundis/alertmanager-syslog)
[![Go Report Card](https://goreportcard.com/badge/github.com/AstroProfundis/alertmanager-syslog)](https://goreportcard.com/report/github.com/AstroProfundis/alertmanager-syslog)

This is an AlertManager receiver that uses the general webhook receiver to send alert notifications to a Syslog server.

NOTE: This repo is still under development and not recommended for production usage.

# Usage
## Run webhook server
Start the server with proper arguments, see `-h` for argument helps. For example, to receive alerts from AlertManager on the same host as webhook server, and forward them to a Syslog server on `192.168.224.224:514`, run with:

```
$ ./bin/alertmanager-syslog -network=udp -syslog="192.168.224.224:514" -host="alert-forwarder.testdomain"
Listening on 0.0.0.0:10514, timeout is 10s
```

To set the hostname (or IP address) section of the message send to Syslog server, use the `-host` argument. If not set, the default local hostname of the server where webhook server is running will be used.

## Configuration
A sample configure file is at [config.yaml](./config.yaml).

The most important config in the configure file is the `mode` section. Currently, `plain` (or `text`), `json` and `custom` are supported value of it.

### Pre-defined modes
Pre-defined modes (`plain`/`test` and `json`) can accept user defined list of labels and annotations from alerts, if the same key name is used for both label and annotation, the value from annotation is used.

All labels and annotations are joint in a `key=value` format and sorted by the key name.

### User defined modes
User defined mode (`custom`) can support custom defined format for messages.

All values defined in `custom.sections` are joint with `custom.delimiter` as delimiter, and the order of sections in the config file is kept.

`custom.sections` defines a list of sections with items:
  - `type`: the type of this section, could be one of "const", "label", "annontation", "time" or "status", where:
      * const is a constitute string set by `value`
      * label is a value from one of the alert's labels
      * annontation is a value from one of the alert's annotations
      * time is the alert start time in UNIX timestamp format (seconds since 1970-01-01 00:00)
      * status is the status of the alert (e.g., firing or resolved)
  - `value`: the constitute string to be used, it is ignored if the type is not "const", if you want to
      keep the section with empty value, use " " (whitespace) for the value.
  - `key`: name of the key for label or annotation, it is ignored if the type is neither "label" nor "annotation"

## Configure AlertManager
Add the following config to your `alertmanager.yaml`:

```
  webhook_configs:
  - url: 'http://<server_ip>:<server_port>/alerts'
```

Where `<server_ip>` and `<server_port>` are the IP address and port the webhook server is listening. In this example, it is `127.0.0.1:10514`

## Fire a testing alert
After the webhook server and AlertManager are all running, use cURL to fire a testing alert:

```
$ curl -X POST "http://localhost:9093/api/v1/alerts" -d '[{
  "status": "firing",
  "labels": {
    "alertname": "testing-alert",
    "service": "some-testing-service",
    "severity":"warning",
    "instance": "test.alert.to.syslog"
  },
  "annotations": {
    "instance": "testalert",
    "summary": "This is a tesing alert."
  },
  "generatorURL": "https://github.com/AstroProfundis/alertmanager-syslog"
}]'
```

And after some secounds, the alert is send to the Syslog server:

```
Oct 28 15:27:53 alert-forwarder.testdomain alertmanager-syslog[6475]: {"alertname":"testing-alert","commonLabels":"testing-alert|test.alert.to.syslog|some-testing-service|warning","severity":"warning","status":"firing","time":"2019-10-28 15:27:43"}
```

If the webhook is started with `-mode=plain`, the message will be:

```
Oct 28 15:31:07 alert-forwarder.testdomain alertmanager-syslog[6876]: status=firing time=2019-10-28 15:30:57 commonLabels=testing-alert|test.alert.to.syslog|some-testing-service|warning alertname=testing-alert severity=warning
```

# Development
## Build
Just run `make`, and find the server binary in `bin/`.

## TODOs
 - Add tests

