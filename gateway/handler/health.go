package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/gateway/service"
	"go.uber.org/zap"
)

type HealthHandler interface {
	Healthcheck(c *gin.Context)
}

type healthHandler struct {
	service service.Service
	logger  *zap.SugaredLogger
}

func NewHealthHandler(logger *zap.SugaredLogger, service service.Service) HealthHandler {
	return &healthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *healthHandler) Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
