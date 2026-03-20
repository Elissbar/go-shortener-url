package shortenerpb

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/handler/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *ShortenerServer) AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("authorization")
		if len(values) > 0 {
			token = values[0]
		}
	}

	userID, err := common.VerifyAuthToken(token, s.Srvc.Config.JWTSecret)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token not found")
	}

	ctx = context.WithValue(ctx, common.UserIDKey, userID)
	return handler(ctx, req)
}
