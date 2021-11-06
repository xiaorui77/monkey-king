package storage

const (
	KeyPrefix = "monkey-king__"
)

type Store interface {
	Visit(url string)
	IsVisited(url string) bool
}
