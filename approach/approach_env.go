// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/evec"
	"github.com/emer/emergent/patgen"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etensor"
)

// Approach implements CS-guided approach to desired outcomes.
// Each location contains a US which satisfies a different drive.
type Approach struct {
	Drives      int                         `desc:"number of different drive-like body states (hunger, thirst, etc), that are satisfied by a corresponding US outcome"`
	CSPerDrive  int                         `desc:"number of different CS sensory cues associated with each US (simplest case is 1 -- one-to-one mapping), presented on a fovea input layer"`
	Locations   int                         `desc:"number of different locations"`
	DistMax     int                         `desc:"maximum distance in time steps to reach the US"`
	NewStateInt int                         `desc:"interval in trials for generating a new state"`
	CSTot       int                         `desc:"total number of CS's = Drives * CSPerDrive"`
	PatSize     evec.Vec2i                  `desc:"size of CS patterns"`
	ActMap      map[string]int              `desc:"action map of action names to indexes"`
	States      map[string]*etensor.Float32 `desc:"named states -- e.g., USs, CSs, etc"`
	PopCode     popcode.OneD                `desc:"population code values, in normalized units"`
	TrgPos      int                         `desc:"target position where Drive US is"`
	Drive       int                         `desc:"current drive state"`
	Dist        int                         `desc:"current distance"`
	Pos         int                         `desc:"current position being looked at"`
	StateCtr    int                         `desc:"count down for generating a new state"`
	Rew         float32                     `desc:"reward"`
	US          int                         `desc:"US is -1 unless consumed at Dist = 0"`
}

func (ev *Approach) Name() string {
	return "Approach"
}

func (ev *Approach) Desc() string {
	return "Approach"
}

// Defaults sets default params
func (ev *Approach) Defaults() {
	ev.Drives = 4
	ev.CSPerDrive = 1
	ev.Locations = 8
	ev.DistMax = 4
	ev.NewStateInt = 4
	ev.PatSize.Set(6, 6)
	ev.PopCode.Defaults()
	ev.PopCode.SetRange(-0.2, 1.2, 0.1)
}

// Config configures the world
func (ev *Approach) Config() {
	ev.CSTot = ev.Drives * ev.CSPerDrive
	ev.ActMap = make(map[string]int)
	ev.ActMap["Forward"] = 0
	ev.ActMap["Left"] = 1
	ev.ActMap["Right"] = 2
	ev.ActMap["Consume"] = 3
	ev.States = make(map[string]*etensor.Float32)
	ev.States["USs"] = etensor.NewFloat32([]int{ev.Locations}, nil, nil)
	ev.States["CSs"] = etensor.NewFloat32([]int{ev.Locations}, nil, nil)
	ev.States["Drive"] = etensor.NewFloat32([]int{ev.Drives}, nil, nil)
	ev.States["US"] = etensor.NewFloat32([]int{ev.Drives + 1}, nil, nil)
	ev.States["CS"] = etensor.NewFloat32([]int{ev.PatSize.Y, ev.PatSize.X}, nil, nil)
	ev.States["Dist"] = etensor.NewFloat32([]int{ev.DistMax, 1}, nil, nil)
	ev.States["Rew"] = etensor.NewFloat32([]int{1, 1}, nil, nil)

	ev.ConfigPats()
	ev.NewState()
}

// ConfigPats generates patterns for CS's
func (ev *Approach) ConfigPats() {
	pats := etensor.NewFloat32([]int{ev.CSTot, ev.PatSize.Y, ev.PatSize.X}, nil, nil)
	patgen.PermutedBinaryMinDiff(pats, 6, 1, 0, 3)
	ev.States["Pats"] = pats
}

func (ev *Approach) Validate() error {
	return nil
}

func (ev *Approach) Init(run int) {
	ev.Config()
}

func (ev *Approach) Counter(scale env.TimeScales) (cur, prv int, changed bool) {
	return 0, 0, false
}

func (ev *Approach) State(el string) etensor.Tensor {
	return ev.States[el]
}

// NewState configures new set of USs in locations
func (ev *Approach) NewState() {
	uss := ev.States["USs"]
	css := ev.States["CSs"]
	for l := 0; l < ev.Locations; l++ {
		us := rand.Intn(ev.Drives)
		cs := rand.Intn(ev.CSPerDrive)
		pat := us*ev.CSPerDrive + cs
		uss.Values[l] = float32(us)
		css.Values[l] = float32(pat)
	}
	ev.StateCtr = ev.NewStateInt
	ev.NewStart()
}

// PatToUS returns US no and CS no from pat no
func (ev *Approach) PatToUS(pat int) (us, cs int) {
	us = pat / ev.CSPerDrive
	cs = pat % ev.CSPerDrive
	return
}

// NewStart starts a new approach run
func (ev *Approach) NewStart() {
	ev.Dist = 1 + rand.Intn(ev.DistMax-1)
	ev.Pos = rand.Intn(ev.Locations)
	ev.TrgPos = rand.Intn(ev.Locations)
	uss := ev.States["USs"]
	ev.Drive = int(uss.Values[ev.TrgPos])
	ev.US = -1
	ev.Rew = 0
	ev.RenderState()
	ev.RenderRewUS()
}

// RenderState renders the current state
func (ev *Approach) RenderState() {
	css := ev.States["CSs"]
	patn := int(css.Values[ev.Pos])
	pats := ev.States["Pats"]
	cs := ev.States["CS"]
	drive := ev.States["Drive"]
	pat := pats.SubSpace([]int{patn})
	cs.CopyFrom(pat)
	drive.SetZeros()
	drive.Values[ev.Drive] = 1
	dist := ev.States["Dist"]
	ev.PopCode.Encode(&dist.Values, float32(ev.Dist), ev.DistMax, false)
}

// RenderRewUS renders reward and US
func (ev *Approach) RenderRewUS() {
	usst := ev.States["US"]
	usst.SetZeros()
	usst.Values[ev.US+1] = 1
	rew := ev.States["Rew"]
	rew.Values[0] = ev.Rew
}

// Step does one step
func (ev *Approach) Step() bool {
	ev.RenderState()
	ev.Rew = 0
	ev.US = -1
	ev.RenderRewUS()
	return true
}

func (ev *Approach) Action(action string, nop etensor.Tensor) {
	_, ok := ev.ActMap[action]
	if !ok {
		fmt.Printf("Action not recognized: %s\n", action)
		return
	}
	uss := ev.States["USs"]
	us := int(uss.Values[ev.Pos])
	switch action {
	case "Forward":
		ev.Dist--
		if ev.Dist <= 0 {
			ev.US = us // still report the us present
			ev.NewStart()
		}
	case "Left":
		ev.Pos--
		if ev.Pos < 0 {
			ev.Pos += ev.Locations
		}
	case "Right":
		ev.Pos++
		if ev.Pos >= ev.Locations {
			ev.Pos -= ev.Locations
		}
	case "Consume":
		if ev.Dist == 0 {
			if us == ev.Drive {
				ev.Rew = 1
			}
			ev.US = us
		}
	}
}
