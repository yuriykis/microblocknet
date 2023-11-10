package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type server struct {
	router  *gin.Engine
	logger  *zap.SugaredLogger
	handler Handler
}

func newServer(logger *zap.SugaredLogger, service *service) *server {
	s := &server{
		router:  gin.Default(),
		logger:  logger,
		handler: newHandler(logger, service),
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.GET("/healthcheck", s.handler.Healthcheck)
	s.router.POST("/block", s.handler.Block)
	s.router.GET("/block/:height", s.handler.Block)
	s.router.GET("/utxo", s.handler.UTXO)
	s.router.POST("/transaction/init", s.handler.InitTransaction)
	s.router.POST("/transaction", s.handler.NewTransaction)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
