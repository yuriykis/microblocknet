package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type server struct {
	router  *gin.Engine
	log     *log.Logger
	handler Handler
}

func newServer() *server {
	s := &server{
		router:  gin.Default(),
		log:     log.New(),
		handler: newHandler(),
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.GET("/healthcheck", s.handler.Healthcheck)
	s.router.POST("/block", s.handler.Block)
	s.router.GET("/block/:height", s.handler.Block)
	s.router.GET("/utxo", s.handler.UTXO)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
