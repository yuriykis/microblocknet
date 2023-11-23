package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/gateway/handler"
	"github.com/yuriykis/microblocknet/gateway/service"
	"go.uber.org/zap"
)

type server struct {
	router *gin.Engine
	logger *zap.SugaredLogger
	bh     handler.BlockHandler
	utxoh  handler.UtxoHandler
	txh    handler.TxHandler
	hh     handler.HealthHandler
}

func newServer(logger *zap.SugaredLogger, service service.Service) *server {
	s := &server{
		router: gin.Default(),
		logger: logger,
		bh:     handler.NewBlockHandler(logger, service),
		utxoh:  handler.NewUtxoHandler(logger, service),
		txh:    handler.NewTxHandler(logger, service),
		hh:     handler.NewHealthHandler(logger, service),
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.GET("/healthcheck", s.hh.Healthcheck)
	s.router.POST("/block", s.bh.Block)
	s.router.GET("/block/:height", s.bh.Block)
	s.router.GET("/utxo", s.utxoh.UTXO)
	s.router.POST("/transaction/init", s.txh.InitTransaction)
	s.router.POST("/transaction", s.txh.NewTransaction)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
