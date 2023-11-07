package service

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

const ttl = time.Second * 10

type ConsulService struct {
	client           *api.Client
	logger           *zap.SugaredLogger
	consulListenAddr string
	connected        bool
}

func NewConsulService(logger *zap.SugaredLogger, consulListenAddr string) *ConsulService {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	return &ConsulService{
		client:           client,
		logger:           logger,
		consulListenAddr: consulListenAddr,
	}
}

func (cs *ConsulService) String() string {
	h, err := cs.getHostname()
	if err != nil {
		cs.logger.Errorf("failed to get hostname: %v", err)
		return ""
	}
	return h + ":" + strings.Split(cs.consulListenAddr, ":")[1]
}

func (cs *ConsulService) getHostname() (string, error) {
	h, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return h, nil
}

func (cs *ConsulService) Start() error {
	cs.logger.Infof("starting %s service", cs)
	ln, err := net.Listen("tcp", cs.consulListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	go cs.tryConnectLoop()
	go cs.update()
	cs.observe()
	return nil
}

func (cs *ConsulService) tryConnectLoop() {
	for {
		time.Sleep(ttl / 2)
		if !cs.connected {
			if err := cs.register(); err != nil {
				cs.logger.Errorf("failed to register service: %v", err)
				cs.connected = false
				continue
			}
			cs.connected = true
		}
	}
}

func (cs *ConsulService) register() error {
	cs.logger.Infof("registering %s service", cs)
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: ttl.String(),
		TTL:                            ttl.String(),
		TLSSkipVerify:                  true,
		CheckID:                        cs.String(),
	}
	consulPortStr := strings.Split(cs.consulListenAddr, ":")[1]
	consulPort, err := strconv.Atoi(consulPortStr)
	if err != nil {
		return err
	}
	service := &api.AgentServiceRegistration{
		ID:      cs.String(),
		Name:    cs.String(),
		Address: strings.Split(cs.consulListenAddr, ":")[0],
		Port:    consulPort,
		Check:   check,
	}
	return cs.client.Agent().ServiceRegister(service)
}

func (cs *ConsulService) observe() error {
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
		cs.logger.Infof("starting %s service observer", cs)
		plan.RunWithConfig("", &api.Config{})
	}()
	return nil
}

func (cs *ConsulService) update() error {
	ticker := time.NewTicker(ttl / 2)
	for {
		select {
		case <-ticker.C:
			cs.logger.Infof("updating %s service", cs)
			if cs.connected {
				if err := cs.client.Agent().UpdateTTL(cs.String(), "", api.HealthPassing); err != nil {
					cs.logger.Errorf("failed to update TTL: %v", err)
					cs.connected = false
					return err
				}
			}
		}
	}
}
