package main

import (
	"context"
	"fmt"
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
	// open a file
	setLogLvl(logger)
}

// main entry point of application
func main() {

	f, err := os.OpenFile("/var/log/testlogrus.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	defer f.Close()

	logger.SetOutput(f)

	logger.Info("Starting MSM Admission Webhook")
	logger.Infof("Version: %v", version)

	// Capture signals and block before exit
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	w := webhook.New(webhook.UseDeps(
		func(d *webhook.Deps) {
			d.Log = logger
		}))

	err = w.Init(ctx)
	if err != nil {
		logger.Fatalf("Could not initialize admission webhook, aborting with error=%s", err)
	}

	var startServerErr = make(chan error)
	go func() {
		startServerErr <- w.Start()
	}()

	select {
	case err := <-startServerErr:
		if ctx.Err() != nil {
			logger.Fatal(err.Error())
		}
	case <-ctx.Done():
		w.Close()
		return
	}

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
