package service

import (
	"log"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

const ttl = time.Second * 10

type ConsulService struct {
	client *api.Client
	logger *zap.SugaredLogger
}

func NewConsulService(logger *zap.SugaredLogger) *ConsulService {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	return &ConsulService{
		client: client,
		logger: logger,
	}
}

func (cs *ConsulService) String() string {
	return "consul"
}

func (cs *ConsulService) Start() error {
	cs.logger.Infof("starting %s service", cs)
	ln, err := net.Listen("tcp", ":10000")
	if err != nil {
		return err
	}
	defer ln.Close()

	if err := cs.register(); err != nil {
		cs.logger.Errorf("failed to register service: %v", err)
		return err
	}
	go cs.update()
	go cs.acceptLoop(ln)
	return nil
}

func (cs *ConsulService) acceptLoop(ln net.Listener) {
	cs.logger.Infof("accepting connections for %s service", cs)
	for {
		time.Sleep(ttl / 2)
		_, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
	}
}

func (cs *ConsulService) register() error {
	cs.logger.Infof("registering %s service", cs)
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: ttl.String(),
		TTL:                            ttl.String(),
		TLSSkipVerify:                  true,
		CheckID:                        "service:" + cs.String(),
	}
	service := &api.AgentServiceRegistration{
		ID:      "service:" + cs.String(),
		Name:    cs.String(),
		Address: "127.0.0.1",
		Port:    10000,
		Check:   check,
	}
	return cs.client.Agent().ServiceRegister(service)
}

func (cs *ConsulService) update() error {
	cs.logger.Infof("updating %s service", cs)
	ticker := time.NewTicker(ttl / 2)
	for {
		select {
		case <-ticker.C:
			if err := cs.client.Agent().UpdateTTL("service:"+cs.String(), "", "pass"); err != nil {
				return err
			}
		}
	}
}
