package main

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	listenAddr := flag.String("listen-addr", ":6000", "The address to listen on for incoming HTTP requests")
	flag.Parse()

	go StartKafkaConsumer()

	if err := Start(*listenAddr); err != nil {
		log.Fatal(err)
	}
}

func Start(listenAddr string) error {
	server := newServer()
	return http.ListenAndServe(listenAddr, server)
}

func StartKafkaConsumer() error {
	kc, err := NewKafkaConsumer([]string{"register_node"})
	if err != nil {
		return err
	}
	kc.Start(newService())
	return nil
}
