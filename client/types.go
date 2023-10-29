package main

type Transaction struct {
	FromAddress []byte
	FromPubKey  []byte
	ToAddress   []byte
	Amount      int
}
