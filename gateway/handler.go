package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/secure"
)

type Handler interface {
	Healthcheck(c *gin.Context)
	Block(c *gin.Context)
	UTXO(c *gin.Context)
	InitTransaction(c *gin.Context)
	NewTransaction(c *gin.Context)
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
	var tx *proto.Transaction
	if c.Request.Method == http.MethodPost {
		var tReq requests.InitTransactionRequest
		if err := c.ShouldBindJSON(&tReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		clientUTXOs, err := h.service.UTXOsByAddress(c.Request.Context(), tReq.FromAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		if clientUTXOs == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "no UTXOs for this address",
			})
		}
		var totalAmount int
		for _, utxo := range clientUTXOs {
			totalAmount += int(utxo.Output.Value)
		}
		if totalAmount < tReq.Amount {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "not enough money",
			})
		}
		prevBlock, err := h.service.BlockByHeight(c.Request.Context(), 0)
		if err != nil {
			log.Fatal(err)
		}
		prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
		txInput := &proto.TxInput{
			PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
			PublicKey:  tReq.FromPubKey,
			OutIndex:   clientUTXOs[0].OutIndex,
		}
		txOutput1 := &proto.TxOutput{
			Value:   int64(tReq.Amount),
			Address: tReq.ToAddress,
		}
		txOutput2 := &proto.TxOutput{
			Value:   int64(totalAmount - tReq.Amount),
			Address: tReq.FromAddress,
		}
		tx = &proto.Transaction{
			Inputs:  []*proto.TxInput{txInput},
			Outputs: []*proto.TxOutput{txOutput1, txOutput2},
		}
	}
	if tx != nil {
		c.JSON(http.StatusOK, requests.InitTransactionResponse{
			Transaction: tx,
		})
	}
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
