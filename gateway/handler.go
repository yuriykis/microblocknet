package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	nodeapi "github.com/yuriykis/microblocknet/node/service/api_client"
)

type Handler interface {
	Healthcheck(c *gin.Context)
	Block(c *gin.Context)
	UTXO(c *gin.Context)
	Transaction(c *gin.Context)
}

type handler struct {
	client nodeapi.Client
}

func newHandler() *handler {
	return &handler{}
}

func (h *handler) Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (h *handler) Block(c *gin.Context) {
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
		b, err := h.client.GetBlockByHeight(c.Request.Context(), height)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, b)
	}
}

func (h *handler) UTXO(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		addressStr := c.Query("address")
		if addressStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "address is required",
			})
		}
		address := []byte(addressStr)
		utxos, err := h.client.GetUTXOsByAddress(c.Request.Context(), address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, utxos)
	}
}

func (h *handler) Transaction(c *gin.Context) {
	return
}
