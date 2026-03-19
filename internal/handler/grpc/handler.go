package shortenerpb

import (
	"context"

	"github.com/Elissbar/go-shortener-url/internal/handler/common"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerServer struct {
	UnimplementedShortenerServiceServer
	srvc *service.Service
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, in *URLShortenRequest) (*URLShortenResponse, error) {
	rq := model.Request{URL: in.GetUrl()}
	userID := ctx.Value(common.UserIDKey).(string)
	resp, err := s.srvc.CreateShortURLJSON(ctx, rq, userID)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error: %s", err.Error())
	}

	result := &URLShortenResponse{}
	result.SetResult(resp.Result)
	return result, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, in *URLExpandRequest) (*URLExpandResponse, error) {
	id := in.GetId()
	userID := ctx.Value(common.UserIDKey).(string)
	url, err := s.srvc.GetShortURL(ctx, id, userID)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error: %s", err.Error())
	}

	result := &URLExpandResponse{}
	result.SetResult(url)
	return result, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, empt *emptypb.Empty) (*UserURLsResponse, error) {
	userID := ctx.Value(common.UserIDKey).(string)
	records, err := s.srvc.GetAllUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error: %s", err.Error())
	}

	var urlData []*URLData
	for _, rec := range records {
		urlData = append(urlData, URLData_builder{ShortUrl: &rec.ShortURL, OriginalUrl: &rec.OriginalURL}.Build())
	}

	result := UserURLsResponse_builder{
		Url: urlData,
	}.Build()
	return result, nil
}
