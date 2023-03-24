/*
 * Copyright (c) 2022 Cisco and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"media-streaming-mesh/msm-admission-webhook/internal/webhook"
)

var (
	version string
	logger  *log.Logger
)

// initializes the logger
func init() {
	logger = log.New()
	// Log as JSON instead of the default ASCII formatter.
	logger.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	logger.SetOutput(os.Stdout)
	setLogLvl(logger)
}

// main entry point of msm-webhook application
func main() {
	logger.Info("Starting MSM Admission Webhook")

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

	err := w.Init(ctx)
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
