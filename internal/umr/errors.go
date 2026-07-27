package umr

import "fmt"

// RPCError is the JSON-RPC error object returned by the UMR portal.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("umr rpc error %d: %s", e.Code, e.Message)
}

// JSON-RPC error codes observed from the UMR portal bundle.
const (
	CodeAccessDenied = -32002
)

// IsAccessDenied reports whether err is a session-expired / wrong-credentials error.
func IsAccessDenied(err error) bool {
	if e, ok := err.(*RPCError); ok {
		return e.Code == CodeAccessDenied
	}
	return false
}
