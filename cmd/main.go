package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cqrs-layout-example/config"
	"cqrs-layout-example/internal/server"

	loader "github.com/NikitaTsaralov/utils/config"
	"github.com/NikitaTsaralov/utils/logger"
)

var (
	exit       = make(chan bool)
	serverImpl *server.Server
)

func main() {
	go listenSignals()
	go start()

	<-exit
	logger.Infof("see you soon...")
}

func start() {
	cfg := &config.Config{}
	loader.LoadYAMLConfig(config.Path, cfg)

	serverImpl = server.NewServer(cfg)
	serverImpl.Start()
}

func listenSignals() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	for {
		s := <-signalChan

		switch s {
		case syscall.SIGHUP:
			logger.Infof("SIGHUP received")

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			err := serverImpl.Stop(ctx)
			if err != nil {
				logger.Fatalf("can't stop file.d with SIGHUP: %s", err.Error())
			}
			cancel()

			start()
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Infof("SIGTERM or SIGINT received")

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			err := serverImpl.Stop(ctx)
			if err != nil {
				logger.Fatalf("can't stop file.d with SIGTERM or SIGINT: %s", err.Error())
			}
			cancel()

			exit <- true
		}
	}
}
