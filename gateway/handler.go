package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
)

type Handler interface {
	Healthcheck(c *gin.Context)
	Block(c *gin.Context)
	UTXO(c *gin.Context)
	InitTransaction(c *gin.Context)
	NewTransaction(c *gin.Context)
	RegisterNode(c *gin.Context)
}

type handler struct {
	service *service
}

func newHandler() *handler {
	return &handler{
		service: newService(),
	}
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
		b, err := h.service.BlockByHeight(c.Request.Context(), height)
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
		utxos, err := h.service.UTXOsByAddress(c.Request.Context(), address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, utxos)
	}
}

func (h *handler) InitTransaction(c *gin.Context) {
	var (
		tx  *proto.Transaction
		err error
	)
	if c.Request.Method == http.MethodPost {
		var tReq requests.InitTransactionRequest
		if err := c.ShouldBindJSON(&tReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		tx, err = h.service.InitTransaction(c.Request.Context(), &Transaction{
			FromAddress: tReq.FromAddress,
			FromPubKey:  tReq.FromPubKey,
			ToAddress:   tReq.ToAddress,
			Amount:      tReq.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
	c.JSON(http.StatusOK, requests.InitTransactionResponse{
		Transaction: tx,
	})
}

func (h *handler) NewTransaction(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		var tReq requests.NewTransactionRequest
		if err := c.ShouldBindJSON(&tReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		t, err := h.service.NewTransaction(c.Request.Context(), tReq.Transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		if t == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "transaction is nil",
			})
		}
		c.JSON(http.StatusOK, requests.NewTransactionResponse{
			Transaction: t,
		})
	}
}

func (h *handler) RegisterNode(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		var nReq requests.RegisterNodeRequest
		if err := c.ShouldBindJSON(&nReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		if err := h.service.NewNode(c.Request.Context(), nReq.Address); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}
