# AlertManager Syslog Receiver

This is an AlertManager receiver that uses the general webhook receiver to send alert notifications to a Syslog server.

NOTE: This repo is still under development and not recommended for production usage.

# Usage
## Run webhook server
Start the server with proper arguments, see `-h` for argument helps.

A sample configure file is at [config.yaml](./config.yaml).

## Configure AlertManager
Add the following config to your `alertmanager.yaml`:

```
  webhook_configs:
  - url: 'http://<server_ip>:<server_port>/alerts'
```

Where `<server_ip>` and `<server_port>` are the IP address and port the webhook server is listening.

# Development
## Build
Just run `make`, and find the server binary in `bin/`.

## TODOs
 - Use a logger instead of `fmt`
 - Support plain text message along side with current JSON format
 - Add an endpoint for Prometheus metrics
 - Add tests

