package types

import "fmt"

func MakeUTXOKey(txHash []byte, index int) string {
	return fmt.Sprintf("%s:%d", txHash, index)
}
