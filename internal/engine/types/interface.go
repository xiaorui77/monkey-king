package types

import "github.com/yougtao/monker-king/internal/view/model"

type Collect interface {
	Visit(url string) error

	GetDataProducer() model.DataProducer
}
