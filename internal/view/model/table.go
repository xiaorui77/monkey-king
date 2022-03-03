package model

import (
	"context"
)

// Tabular 表格形式
type Tabular interface {
	Start()
	Watch(ctx context.Context) error
}

type Table struct {
	data DataProducer
}

func (t *Table) Watch(ctx context.Context) error {
	if err := t.refresh(ctx); err != nil {
		return err
	}
	go t.update(ctx)
	return nil
}

func (t *Table) refresh(ctx context.Context) error {
	// todo
	return nil
}

func (t *Table) update(ctx context.Context) {

}
