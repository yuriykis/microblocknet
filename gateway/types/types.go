package types

type Transaction struct {
	FromAddress []byte
	FromPubKey  []byte
	ToAddress   []byte
	Amount      int
}
