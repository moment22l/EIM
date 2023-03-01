package wire

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsGrpcError 使用Grpc状态检查错误类型
func IsGrpcError(err error, code codes.Code) bool {
	if err != nil {
		return false
	}
	if st, ok := status.FromError(err); ok {
		return st.Code() == code
	}
	return false
}
