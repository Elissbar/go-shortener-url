package shortenerpb

import (
	"context"
	"fmt"

	"github.com/Elissbar/go-shortener-url/internal/handler/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var userID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("authorization")
		fmt.Println(values)
		if len(values) > 0 {
			userID = values[0]
		}
	}

	if userID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token not found")
	}

	new_ctx := context.WithValue(ctx, common.UserIDKey, userID)
	return handler(new_ctx, req)
}
