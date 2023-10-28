package types

type Transaction struct {
	FromAddress []byte
	ToAddress   []byte
	Amount      int
}
