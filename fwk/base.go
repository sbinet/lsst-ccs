package fwk

import "github.com/gonuts/logger"

type Base struct {
	*logger.Logger
}

func NewBase(name string) *Base {
	return &Base{
		Logger: logger.New(name),
	}
}
