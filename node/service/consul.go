package service

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

const ttl = time.Second * 10

type ConsulService struct {
	client     *api.Client
	logger     *zap.SugaredLogger
	nodeName   string
	listenAddr string
}

func NewConsulService(logger *zap.SugaredLogger, nodeName string, listenAddr string) *ConsulService {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	return &ConsulService{
		client:     client,
		logger:     logger,
		nodeName:   nodeName,
		listenAddr: listenAddr,
	}
}

func (cs *ConsulService) String() string {
	return "node" + cs.nodeName
}

func (cs *ConsulService) Start() error {
	cs.logger.Infof("starting %s service", cs)
	ln, err := net.Listen("tcp", cs.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if err := cs.register(); err != nil {
		cs.logger.Errorf("failed to register service: %v", err)
		return err
	}
	go cs.update()
	return nil
}

func (cs *ConsulService) register() error {
	cs.logger.Infof("registering %s service", cs)
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: ttl.String(),
		TTL:                            ttl.String(),
		TLSSkipVerify:                  true,
		CheckID:                        cs.String(),
	}
	consulPortStr := strings.Split(cs.listenAddr, ":")[1]
	consulPort, err := strconv.Atoi(consulPortStr)
	if err != nil {
		return err
	}
	service := &api.AgentServiceRegistration{
		ID:      cs.String(),
		Name:    cs.String(),
		Address: strings.Split(cs.listenAddr, ":")[0],
		Port:    consulPort,
		Check:   check,
	}

	query := map[string]any{
		"type":        "service",
		"service":     cs.String(),
		"passingonly": true,
	}
	plan, err := watch.Parse(query)
	if err != nil {
		return err
	}
	plan.HybridHandler = func(idx watch.BlockingParamVal, result any) {
		switch result.(type) {
		case []*api.ServiceEntry:
			for _, entry := range result.([]*api.ServiceEntry) {
				cs.logger.Infof("service %s updated: %v", cs, entry)
			}
		default:
			cs.logger.Infof("service %s updated", cs)
		}
	}
	go func() {
		plan.RunWithConfig("", &api.Config{})
	}()

	return cs.client.Agent().ServiceRegister(service)
}

func (cs *ConsulService) update() error {
	cs.logger.Infof("updating %s service", cs)
	ticker := time.NewTicker(ttl / 2)
	for {
		select {
		case <-ticker.C:
			if err := cs.client.Agent().UpdateTTL(cs.String(), "", api.HealthPassing); err != nil {
				return err
			}
		}
	}
}
