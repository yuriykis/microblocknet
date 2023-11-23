package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/gateway/service"
	"github.com/yuriykis/microblocknet/gateway/types"
	"go.uber.org/zap"
)

type TxHandler interface {
	InitTransaction(c *gin.Context)
	NewTransaction(c *gin.Context)
}

type txHandler struct {
	service service.Service
	logger  *zap.SugaredLogger
}

func NewTxHandler(logger *zap.SugaredLogger, service service.Service) TxHandler {
	return &txHandler{
		service: service,
		logger:  logger,
	}
}

func (h *txHandler) InitTransaction(c *gin.Context) {
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
		tx, err = h.service.InitTransaction(c.Request.Context(), &types.Transaction{
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

func (h *txHandler) NewTransaction(c *gin.Context) {
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
