package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	webhook "github.com/AstroProfundis/alertmanager-syslog/pkg"
	"github.com/AstroProfundis/alertmanager-syslog/pkg/version"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configFile   string
	listenAddr   string
	syslogAddr   string
	network      string
	tag          string
	noPid        bool
	timeout      int
	host         string
	printVersion bool
)

func init() {
	flag.StringVar(&configFile, "config", "", "Config file of labels and annotations to send to syslog.")
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:10514", "Address and port of the webhook to receive messages from AlertManager.")
	flag.StringVar(&syslogAddr, "syslog", "127.0.0.1:514", "Address and port of the Syslog server to send messages.")
	flag.StringVar(&network, "network", "", "(tcp or udp): send messages to the syslog server using UDP or TCP. If not set, connect to the local syslog server.")
	flag.StringVar(&tag, "tag", "alertmanager-syslog", "The tag used in syslog messages.")
	flag.BoolVar(&noPid, "no-pid", false, "Exclude PID of the webhook server from syslog message.")
	flag.IntVar(&timeout, "timeout", 10, "Timeout when serving and sending requests, in seconds.")
	flag.StringVar(&host, "host", "", "Hostname or IP address of the log source, if not set, default local hostname will be used.")
	flag.BoolVar(&printVersion, "V", false, "Show version and quit.")
	flag.BoolVar(&printVersion, "version", false, "Show version and quit.")
	flag.Parse()
}

func main() {
	ver := version.NewVersion()
	if printVersion {
		fmt.Println(ver)
		os.Exit(0)
	}
	glog.Infof("Start alertmanager-syslog %s as %s", ver, tag)

	cfg, err := webhook.LoadConfig(configFile)
	if err != nil {
		glog.Fatalf("Failed to load config file: %v", err)
		os.Exit(1)
	}

	s, err := webhook.New(&webhook.ServerCfg{
		ListenAddr: listenAddr,
		SyslogAddr: syslogAddr,
		Network:    network,
		Tag:        tag,
		NoPid:      noPid,
		Timeout:    timeout,
		Hostname:   host,
		Config:     cfg,
	})
	if err != nil {
		glog.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}

	http.HandleFunc("/alerts", s.HandleAlert)
	http.HandleFunc("/version", s.ShowVersion)
	http.Handle("/metrics", promhttp.Handler())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			s.Close()
			glog.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		glog.Infof("Got signal [%v], exiting...\n", sig)
		s.Close()
	}()

	wg.Wait()
	glog.Flush()
}
