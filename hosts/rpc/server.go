package rpc

import (
	"fmt"

	"github.com/jrapoport/gothic/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server represents an gRPC server.
type Server struct {
	*core.Server
}

// NewServer creates a gRPC Server.
func NewServer(s *core.Server) *Server {
	return &Server{s}
}

/*
func (s *Server) Response(in interface{}, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return s.RPCError(codes.Internal, err)
	}
	err = json.Unmarshal(b, out)
	return s.RPCError(codes.Internal, err)
}
*/

// RPCError wraps an rpc error code.
func (s *Server) RPCError(c codes.Code, err error) error {
	if c == codes.OK {
		return nil
	}
	msg := statusText(c)
	if err != nil {
		msg = err.Error()
	}
	err = status.Error(c, msg)
	s.Error(err)
	return err
}

// RPCErrorf wraps an rpc error code formatted according to a format specifier.
func (s *Server) RPCErrorf(c codes.Code, format string, a ...interface{}) error {
	return s.RPCError(c, fmt.Errorf(format, a...))
}

func statusText(c codes.Code) string {
	switch c {
	case codes.OK:
		return "ok"
	case codes.Canceled:
		return "cancelled"
	case codes.Unknown:
		return "unknown"
	case codes.InvalidArgument:
		return "invalid argument"
	case codes.DeadlineExceeded:
		return "deadline exceeded"
	case codes.NotFound:
		return "not found"
	case codes.AlreadyExists:
		return "already exists"
	case codes.PermissionDenied:
		return "permission denied"
	case codes.ResourceExhausted:
		return "resource exhausted"
	case codes.FailedPrecondition:
		return "failed precondition"
	case codes.Aborted:
		return "aborted"
	case codes.OutOfRange:
		return "out of range"
	case codes.Unimplemented:
		return "unimplemented"
	case codes.Internal:
		return "internal"
	case codes.Unavailable:
		return "unavailable"
	case codes.DataLoss:
		return "data loss"
	case codes.Unauthenticated:
		return "unauthenticated"
	default:
		return "error"
	}
}
