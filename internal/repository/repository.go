package repository

type Storage interface {
	Save(token, url string) error
	Get(token string) (string, bool)
	Close() error
}
