package interfaces

type WSManager interface {
	SetUrl(url string)
	Connect() error
	Reconnect(url string) error
	Close() error
	Send(data []byte)
	Receive() <-chan []byte
	Errors() <-chan error
	IsConnected() bool
}
