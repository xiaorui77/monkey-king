package types

type Collect interface {
	Visit(url string) error
}
