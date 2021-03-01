package main  //agent

import (
	"github.com/emer/etable/etensor"
)

type BaseAgent interface {
	Step(state map[string]etensor.Tensor) map[string]etensor.Tensor
	Defaults()
	Name() string
}