package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/media-streaming-mesh/msm-admission-webhook/internal/webhook"
	log "github.com/sirupsen/logrus"
)

var (
	version string
	logger  *log.Logger
)

// initializes the logger
func init() {
	logger = log.New()
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	setLogLvl(logger)
}

// main entry point of application
func main() {
	logger.Info("Starting SM Admission Webhook")
	logger.Infof("Version: %v", version)

	// Capture signals and block before exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)

	w := webhook.New(webhook.UseDeps(
		func(d *webhook.Deps) {
			d.Log = logger
		}))

	err := w.Init()
	if err != nil {
		w.Log.Fatalf("Could not initialize admission webhook, aborting with error=%s", err)
	}

	go func() {
		err := w.Start()
		if err != nil {
			w.Log.Fatalf("Could not start webhook server, aborting with error=%s", err)
		}
	}()

	<-c
	// cleanup
	w.Close()
}

// sets the log level of the logger
func setLogLvl(l *log.Logger) {
	logLevel := os.Getenv("LOG_LVL")

	switch logLevel {
	case "DEBUG":
		l.SetLevel(log.DebugLevel)
	case "WARN":
		l.SetLevel(log.WarnLevel)
	case "INFO":
		l.SetLevel(log.InfoLevel)
	case "ERROR":
		l.SetLevel(log.ErrorLevel)
	case "TRACE":
		l.SetLevel(log.TraceLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		l.SetLevel(log.WarnLevel)
	}

}
