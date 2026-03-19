package shortenerpb

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/service"
)

type ShortenerServer struct {
	UnimplementedShortenerServiceServer
	srvc *service.Service
}

// ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error)
func (s *ShortenerServer) ShortenURL(ctx context.Context, in *URLShortenRequest) (*URLShortenResponse, error) {
	// rq := model.Request{URL: in.GetUrl()}
	// s.srvc.CreateShortURLJSON(ctx, rq, userID)
	return nil, nil
}
