package types

import "github.com/xiaorui77/monker-king/internal/view/model"

type Collect interface {
	Visit(url string) error

	GetDataProducer() model.DataProducer
}
