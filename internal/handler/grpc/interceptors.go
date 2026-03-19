package shortenerpb

import (
	context "context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var user_id string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("authorization")
		fmt.Println(values)
		if len(values) > 0 {
			user_id = values[0]
		}
	}

	if user_id == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token not found")
	}

	new_ctx := context.WithValue(ctx, "user_id", user_id)
	return handler(new_ctx, req)
}
