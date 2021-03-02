package main

import (
	"math/rand"

	"github.com/emer/etable/etensor"
	"github.com/emer/emergent/popcode"
	"github.com/goki/mat32"
)

// Agent for selecting random saccade plans for testing
type RandomAgent struct {
	Nm            string           `desc: "Agent name"`
	SacGenMax     float32          `desc:"maximum saccade size"`
	SacPlanFloat  mat32.Vec2       `desc:"eye movement plan, as an offset in world coordinates"`
	SacPlan       etensor.Float32  `desc: saccade plan as an offset in world coordinates as a 2D Gaussian bump population code`
	SacPop        popcode.TwoD     `desc:"2d population code for gaussian bump rendering of saccade plan / execution"`
}

func (ag *RandomAgent) Name() string {
	return ag.Nm
}

func (ag *RandomAgent) Defaults() {
	ag.Nm = "RandomAgent"
	ag.SacGenMax = 0.4

	ag.SacPop.Defaults()
	ag.SacPop.Min.Set(-0.45, -0.45)
	ag.SacPop.Max.Set(0.45, 0.45)

	ag.SacPlan.SetShape([]int{11, 11}, nil, nil)
}

func (ag *RandomAgent) Step(state map[string]etensor.Tensor) map[string]etensor.Tensor {  //Float32 {  // TODO state/action structs, func Step(s *StateStruct) (a ActionStruct) {
	ag.SacPlanFloat.X = 2.0 * (rand.Float32() -  0.5) * ag.SacGenMax
	ag.SacPlanFloat.Y = 2.0 * (rand.Float32() -  0.5) * ag.SacGenMax
	ag.SacPop.Encode(&ag.SacPlan, ag.SacPlanFloat, popcode.Set)
	action := map[string]etensor.Tensor {"SacPlan": etensor.NewFloat32Shape(etensor.NewShape([]int{11, 11}, nil, nil), ag.SacPlan.Values)}
	return action
}
