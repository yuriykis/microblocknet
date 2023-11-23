package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/gateway/service"
	"go.uber.org/zap"
)

type BlockHandler interface {
	Block(c *gin.Context)
}

type blockHandler struct {
	service service.Service
	logger  *zap.SugaredLogger
}

func NewBlockHandler(logger *zap.SugaredLogger, service service.Service) BlockHandler {
	return &blockHandler{
		service: service,
		logger:  logger,
	}
}

func (h *blockHandler) Block(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		heightStr := c.Param("height")
		if heightStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "height is required",
			})
		}
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "height must be an integer",
			})
		}
		b, err := h.service.BlockByHeight(c.Request.Context(), height)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, b)
	}
}
