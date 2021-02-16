package ws

import "fmt"

type ConnectionErrorType int

const (
	NodeError ConnectionErrorType = iota
	PortPoolError
	UserCancellationError
)

type ConnectionError struct {
	Err  error
	Type ConnectionErrorType
}

func (r *ConnectionError) Error() string {
	return fmt.Sprintf("connection error type %d: err %v", r.Type, r.Err)
}

func (r *ConnectionError) IsNodeError() bool {
	return r.Type == NodeError
}
