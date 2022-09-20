package grpc_errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InvalidArgument
func InvalidArgumentError(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

// Internal
func InternalError(msg string) error {
	return status.Error(codes.Internal, msg)
}

// NotFound
func NotFoundError(msg string) error {
	return status.Error(codes.NotFound, msg)
}
