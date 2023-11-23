package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/gateway/service"
	"go.uber.org/zap"
)

type UtxoHandler interface {
	UTXO(c *gin.Context)
}

type uxtoHandler struct {
	service service.Service
	logger  *zap.SugaredLogger
}

func NewUtxoHandler(logger *zap.SugaredLogger, service service.Service) UtxoHandler {
	return &uxtoHandler{
		service: service,
		logger:  logger,
	}
}

func (h *uxtoHandler) UTXO(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		addressStr := c.Query("address")
		if addressStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "address is required",
			})
		}
		address := []byte(addressStr)
		utxos, err := h.service.UTXOsByAddress(c.Request.Context(), address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, utxos)
	}
}
