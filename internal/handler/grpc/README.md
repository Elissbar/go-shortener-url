Запросы для тестирования с помощью grpcurl:

grpcurl -proto internal/handler/grpc/shortener.proto -import-path internal/handler/grpc -plaintext -d '{"url": "https://esdfdsfxampledomainsfasdfasf.com"}' -H "authorization: token" localhost:3200 shortener.ShortenerService/ShortenURL

grpcurl -proto internal/handler/grpc/shortener.proto -import-path internal/handler/grpc -plaintext -d '{"id": "o-Gd5x8KHUY"}' -H "authorization: token" localhost:3200 shortener.ShortenerService/ExpandURL

grpcurl -proto internal/handler/grpc/shortener.proto -import-path internal/handler/grpc -plaintext -plaintext -H "authorization: token" localhost:3200 shortener.ShortenerService/ListUserURLs