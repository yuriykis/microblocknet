package node

type TransportServer interface {
	Start() error
	Stop() error
}
