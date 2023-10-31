package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/secure"
	nodeapi "github.com/yuriykis/microblocknet/node/service/api_client"
)

type service struct {
	nodeapi nodeapi.Client
}

func newService() *service {
	return &service{
		nodeapi: nodeapi.NewHTTPClient("http://localhost:4000"),
	}
}

func (s *service) GetBlockByHeight(c *gin.Context, height int) (requests.GetBlockByHeightResponse, error) {
	return s.nodeapi.GetBlockByHeight(c, height)
}

func (s *service) GetUTXOsByAddress(c *gin.Context, address []byte) (*requests.GetUTXOsByAddressResponse, error) {
	return s.nodeapi.GetUTXOsByAddress(c, address)
}

func (s *service) InitTransaction(
	c *gin.Context,
	tReq requests.InitTransactionRequest,
) (*proto.Transaction, error) {
	clientUTXOs, err := s.nodeapi.GetUTXOsByAddress(c.Request.Context(), tReq.FromAddress)
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
	for _, utxo := range clientUTXOs.UTXOs {
		totalAmount += int(utxo.Output.Value)
	}
	if totalAmount < tReq.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not enough money",
		})
	}
	prevBlockRes, err := s.nodeapi.GetBlockByHeight(c.Request.Context(), 0)
	if err != nil {
		log.Fatal(err)
	}
	prevBlock := prevBlockRes.Block
	prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
	txInput := &proto.TxInput{
		PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
		PublicKey:  tReq.FromPubKey,
		OutIndex:   clientUTXOs.UTXOs[0].OutIndex,
	}
	txOutput1 := &proto.TxOutput{
		Value:   int64(tReq.Amount),
		Address: tReq.ToAddress,
	}
	txOutput2 := &proto.TxOutput{
		Value:   int64(totalAmount - tReq.Amount),
		Address: tReq.FromAddress,
	}
	return &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}, nil
}

func (s *service) NewTransaction(
	c *gin.Context,
	tReq requests.NewTransactionRequest,
) (requests.NewTransactionResponse, error) {
	return s.nodeapi.NewTransaction(c, tReq)
}
