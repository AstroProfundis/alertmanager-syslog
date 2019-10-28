# AlertManager Syslog Receiver

This is an AlertManager receiver that uses the general webhook receiver to send alert notifications to a Syslog server.

NOTE: This repo is still under development and not recommended for production usage.

# Usage
## Run webhook server
Start the server with proper arguments, see `-h` for argument helps. For example, to receive alerts from AlertManager on the same host as webhook server, and forward them to a Syslog server on `192.168.224.224:514`, run with:

```
$ ./bin/alertmanager-syslog -network=udp -syslog="192.168.224.224:514" -host="alert-forwarder.testdomain"
Listening on 0.0.0.0:10514, timeout is 10s
```

To send plain text (rather than JSON) to Syslog server, add `-mode=plain` to the command arguments.

To set the hostname (or IP address) section of the message send to Syslog server, use the `-host` argument. If not set, the default local hostname of the server where webhook server is running will be used.

A sample configure file is at [config.yaml](./config.yaml).

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
 - Support plain text message along side with current JSON format
 - Add an endpoint for Prometheus metrics
 - Add tests

