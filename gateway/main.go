package main

import (
	"flag"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	listenAddr := flag.String("listen-addr", ":6000", "The address to listen on for incoming HTTP requests")
	flag.Parse()

	logger := makeLogger()

	go StartKafkaConsumer(logger)

	if err := Start(*listenAddr, logger); err != nil {
		log.Fatal(err)
	}
}

func Start(listenAddr string, logger *zap.SugaredLogger) error {
	server := newServer(logger)
	return http.ListenAndServe(listenAddr, server)
}

func StartKafkaConsumer(logger *zap.SugaredLogger) error {
	kc, err := NewKafkaConsumer([]string{"register_node"})
	if err != nil {
		return err
	}
	kc.Start(newService(logger))
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
