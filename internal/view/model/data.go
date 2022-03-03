package model

type DataProducer interface {
	GetRows() []interface{}
}
