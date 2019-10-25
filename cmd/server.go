package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	webhook "github.com/AstroProfundis/alertmanager-syslog/pkg"
)

var (
	listen  string
	syslog  string
	timeout int
)

func initArgs() {
	flag.StringVar(&listen, "listen", "0.0.0.0:10514", "Address and port of the webhook to receive messages from AlertManager.")
	flag.StringVar(&syslog, "syslog", "127.0.0.1:514", "Address and port of the Syslog server to send messages.")
	flag.IntVar(&timeout, "timeout", 10, "Timeout when serving and sending requests, in seconds.")
	flag.Parse()
}

func init() {
	initArgs()
}

func main() {
	timeoutSec := time.Second * time.Duration(timeout)

	http.HandleFunc("/alerts", webhook.HandleAlert)

	s := &http.Server{
		Addr:         listen,
		ReadTimeout:  timeoutSec,
		WriteTimeout: timeoutSec,
	}
	fmt.Printf("Listening on %s, timeout is %v\n", listen, timeoutSec)
	s.ListenAndServe()
}
