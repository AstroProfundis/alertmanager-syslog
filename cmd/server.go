package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	webhook "github.com/AstroProfundis/alertmanager-syslog/pkg"
	"github.com/golang/glog"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configFile string
	listenAddr string
	syslogAddr string
	network    string
	timeout    int
	host       string
)

func init() {
	flag.StringVar(&configFile, "config", "", "Config file of labels and annotations to send to syslog.")
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:10514", "Address and port of the webhook to receive messages from AlertManager.")
	flag.StringVar(&syslogAddr, "syslog", "127.0.0.1:514", "Address and port of the Syslog server to send messages.")
	flag.StringVar(&network, "network", "", "(tcp or udp): send messages to the syslog server using UDP or TCP. If not set, connect to the local syslog server.")
	flag.IntVar(&timeout, "timeout", 10, "Timeout when serving and sending requests, in seconds.")
	flag.StringVar(&host, "host", "", "Hostname or IP address of the log source, if not set, default local hostname will be used.")
	flag.Parse()
}

func main() {
	cfg, err := webhook.LoadConfig(configFile)
	if err != nil {
		glog.Fatalf("Failed to load config file: %v", err)
		os.Exit(1)
	}

	s, err := webhook.New(&webhook.ServerCfg{
		ListenAddr: listenAddr,
		SyslogAddr: syslogAddr,
		Network:    network,
		Timeout:    timeout,
		Hostname:   host,
		Config:     cfg,
	})
	if err != nil {
		glog.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
	defer s.Close()

	http.HandleFunc("/alerts", s.HandleAlert)
	http.Handle("/metrics", promhttp.Handler())

	hc := healthcheck.NewMetricsHandler(prometheus.DefaultRegisterer, "alertmanager-syslog")
	http.HandleFunc("/live", hc.LiveEndpoint)
	http.HandleFunc("/ready", hc.ReadyEndpoint)

	go s.ListenAndServe()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		sig := <-sc
		glog.Infof("Got signal [%v], exiting...\n", sig)
		wg.Done()
	}()
	wg.Wait()
	glog.Flush()
}
