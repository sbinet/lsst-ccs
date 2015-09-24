package fwk

import (
	"time"

	"github.com/gonuts/logger"
)

type Base struct {
	*logger.Logger
	ticker *time.Ticker
}

func NewBase(name string) *Base {
	return &Base{
		Logger: logger.New(name),
		ticker: time.NewTicker(1 * time.Second),
	}
}

func (b *Base) Tick() <-chan time.Time {
	return b.ticker.C
}
