package main

import (
	"flag"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/gateway/network"
	"github.com/yuriykis/microblocknet/gateway/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	listenAddr := flag.String("listen-addr", ":6000", "The address to listen on for incoming HTTP requests")
	flag.Parse()

	logger := makeLogger()
	network := network.New(logger)
	service := service.New(logger, network)

	go StartKafkaConsumer(logger, network)

	if err := Start(*listenAddr, logger, service); err != nil {
		log.Fatal(err)
	}
}

func Start(listenAddr string, logger *zap.SugaredLogger, service service.Service) error {
	server := newServer(logger, service)
	return http.ListenAndServe(listenAddr, server)
}

func StartKafkaConsumer(logger *zap.SugaredLogger, n network.Networker) error {
	kc, err := NewKafkaConsumer([]string{"register_node"}, logger)
	if err != nil {
		return err
	}
	kc.Start(n)
	return nil
}

func makeLogger() *zap.SugaredLogger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
