// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/evec"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ints"
)

// FWorld is a flat-world grid-based environment
type FWorld struct {
	Nm        string                      `desc:"name of this environment"`
	Dsc       string                      `desc:"description of this environment"`
	Size      evec.Vec2i                  `desc:"size of 2D world"`
	PatSize   evec.Vec2i                  `desc:"size of patterns for mats, acts"`
	World     etensor.Int                 `desc:"2D grid world, each cell is a material (mat)"`
	Mats      []string                    `desc:"list of materials in the world, 0 = empty.  Any superpositions of states (e.g., CoveredFood) need to be discretely encoded, can be transformed through action rules"`
	MatMap    map[string]int              `desc:"map of material name to index stored in world cell"`
	MatColors map[string]string           `desc:"color strings for different material types, for view"`
	MatPats   map[string]*etensor.Float32 `desc:"patterns for each material"`
	Acts      []string                    `desc:"list of actions: starts with: Stay, Left, Right, Forward, Back, then extensible"`
	ActMap    map[string]int              `desc:"action map of action names to indexes"`
	ActPats   map[string]*etensor.Float32 `desc:"patterns for each action"`
	Inters    []string                    `desc:"list of interoceptive body states, represented as pop codes"`
	InterMap  map[string]int              `desc:"map of interoceptive state names to indexes"`
	Params    map[string]float32          `desc:"map of optional interoceptive and world-dynamic parameters -- cleaner to store in a map"`
	AngInc    int                         `desc:"angle increment for rotation, in degrees -- defaults to 15"`
	FOV       int                         `desc:"field of view in degrees, e.g., 180, must be even multiple of AngInc"`
	PopSize   int                         `desc:"number of units in population codes"`
	PopCode   popcode.OneD                `desc:"population code values, in normalized units"`

	// current state below (params above)
	Pos         evec.Vec2i                  `inactive:"+" desc:"current location of agent"`
	Angle       int                         `inactive:"+" desc:"current angle, in degrees"`
	InterStates map[string]float32          `inactive:"+" desc:"floating point value of internal states -- dim of Inters"`
	States      map[string]*etensor.Float32 `desc:"rendered state tensors-- extensible map"`
	Run         env.Ctr                     `view:"inline" desc:"current run of model as provided during Init"`
	Epoch       env.Ctr                     `view:"inline" desc:"increments over arbitrary fixed number of trials, for general stats-tracking"`
	Trial       env.Ctr                     `view:"inline" desc:"increments for each step of world, loops over epochs -- for general stats-tracking independent of env state"`
	Event       env.Ctr                     `view:"arbitrary counter for steps within a scene"`
	Scene       env.Ctr                     `view:"arbitrary counter incrementing over a coherent sequence of events: e.g., approaching food"`
	Episode     env.Ctr                     `view:"arbitrary counter incrementing over scenes within larger episode: feeding, drinking, exploring, etc"`
}

func (ev *FWorld) Name() string { return ev.Nm }
func (ev *FWorld) Desc() string { return ev.Dsc }

// Config configures the world
func (ev *FWorld) Config(ntrls int) {
	ev.Nm = "Demo"
	ev.Dsc = "Example world with basic food / water / eat / drink actions"
	ev.Mats = []string{"Empty", "Wall", "Food", "Water"}
	ev.MatColors = map[string]string{
		"Empty": "lightgrey", "Wall": "black", "Food": "orange", "Water": "blue",
	}
	ev.Acts = []string{"Stay", "Left", "Right", "Forward", "Back", "Eat", "Drink"}
	ev.Inters = []string{"Energy", "Hydra", "FoodRew", "WaterRew"}

	ev.Params = make(map[string]float32)

	ev.Params["StepCost"] = 0.01   // additional decrement due to stepping forward
	ev.Params["TimeCost"] = 0.01   // decrement due to existing for 1 unit of time, in energy and hydration
	ev.Params["StepCost"] = 0.01   // additional decrement due to stepping forward
	ev.Params["RotCost"] = 0.001   // additional decrement due to rotating one step
	ev.Params["EatCost"] = 0.001   // additional decrement in hydration due to eating
	ev.Params["DrinkCost"] = 0.001 // additional decrement in energy due to drinking
	ev.Params["EatVal"] = 0.1      // increment in energy due to eating one unit of food
	ev.Params["DrinkVal"] = 0.1    // increment in hydration due to drinking one unit of water

	ev.Size.Set(100, 100)
	ev.PatSize.Set(5, 5)
	ev.AngInc = 15
	ev.FOV = 180
	ev.PopSize = 12
	ev.PopCode.Defaults()
	ev.PopCode.SetRange(-0.2, 1.2, 0.1)

	ev.Trial.Max = ntrls

	ev.ConfigImpl()
}

// ConfigImpl does the automatic parts of configuration
// generally does not require editing
func (ev *FWorld) ConfigImpl() {
	ev.World.SetShape([]int{ev.Size.Y, ev.Size.X}, nil, []string{"Y", "X"})

	ev.States = make(map[string]*etensor.Float32)

	dv := &etensor.Float32{}
	nang := (ev.FOV / ev.AngInc) + 1
	dv.SetShape([]int{1, nang, ev.PopSize, 1}, nil, []string{"1", "Angle", "Pop", "1"})
	ev.States["DepthView"] = dv

	fv := &etensor.Float32{}
	fv.SetShape([]int{ev.PatSize.Y, ev.PatSize.X}, nil, []string{"Y", "X"})
	ev.States["Fovea"] = fv

	ps := &etensor.Float32{}
	ps.SetShape([]int{1, 4, 2, 1}, nil, []string{"1", "Pos", "OnOff", "1"})
	ev.States["ProxSoma"] = ps

	vs := &etensor.Float32{}
	ps.SetShape([]int{1, ev.PopSize}, nil, []string{"1", "Pop"})
	ev.States["Vestibular"] = vs

	is := &etensor.Float32{}
	ps.SetShape([]int{1, len(ev.Inters), ev.PopSize, 1}, nil, []string{"1", "Inters", "Pop", "1"})
	ev.States["Inters"] = is

	ev.MatMap = make(map[string]int, len(ev.Mats))
	for i, m := range ev.Mats {
		ev.MatMap[m] = i
	}
	ev.ActMap = make(map[string]int, len(ev.Acts))
	for i, m := range ev.Acts {
		ev.ActMap[m] = i
	}
	ev.InterMap = make(map[string]int, len(ev.Inters))
	for i, m := range ev.Inters {
		ev.InterMap[m] = i
	}
	ev.InterStates = make(map[string]float32, len(ev.Inters))
	for _, m := range ev.Inters {
		ev.InterStates[m] = 0
	}

	ev.Run.Scale = env.Run
	ev.Epoch.Scale = env.Epoch
	ev.Trial.Scale = env.Trial
	ev.Event.Scale = env.Event
	ev.Scene.Scale = env.Scene
	ev.Episode.Scale = env.Episode
}

func (ev *FWorld) Validate() error {
	if ev.Size.IsNil() {
		return fmt.Errorf("FWorld: %v has size == 0 -- need to Config", ev.Nm)
	}
	return nil
}

func (ev *FWorld) State(element string) etensor.Tensor {
	return ev.States[element]
}

// String returns the current state as a string
func (ev *FWorld) String() string {
	return fmt.Sprintf("Pos_%d_%d", ev.Pos.X, ev.Pos.Y)
}

// Init is called to restart environment
func (ev *FWorld) Init(run int) {
	ev.Run.Init()
	ev.Epoch.Init()
	ev.Trial.Init()
	ev.Event.Init()
	ev.Scene.Init()
	ev.Episode.Init()

	ev.Run.Cur = run
	ev.Trial.Cur = -1 // init state -- key so that first Step() = 0
	ev.Event.Cur = -1

	ev.Pos = ev.Size.DivScalar(2) // start in middle -- could be random..
	ev.Angle = 0
	ev.InterStates["Energy"] = 1
	ev.InterStates["Hydra"] = 1
	ev.InterStates["FoodRew"] = 0
	ev.InterStates["WaterRew"] = 0
}

// SaveWorld saves the world to a tsv file with empty string for empty cells
func (ev *FWorld) SaveWorld(filename gi.FileName) error {
	fp, err := os.Create(string(filename))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer fp.Close()
	bw := bufio.NewWriter(fp)
	for y := 0; y < ev.Size.Y; y++ {
		for x := 0; x < ev.Size.X; x++ {
			mat := ev.World.Value([]int{y, x})
			ms := ev.Mats[mat]
			if ms == "Empty" {
				ms = ""
			}
			bw.WriteString(ms + "\t")
		}
		bw.WriteString("\n")
	}
	bw.Flush()
	return nil
}

// OpenWorld loads the world from a tsv file with empty string for empty cells
func (ev *FWorld) OpenWorld(filename gi.FileName) error {
	fp, err := os.Open(string(filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer fp.Close()
	ev.World.SetZeros()
	scan := bufio.NewScanner(fp)
	for y := 0; y < ev.Size.Y; y++ {
		ln := scan.Bytes()
		sz := len(ln)
		if sz == 0 {
			break
		}
		sp := bytes.Split(ln, []byte("\t"))
		sz = ints.MinInt(sz, len(sp)-1)
		for x := 0; x < ev.Size.X; x++ {
			ms := string(sp[x])
			if ms == "" {
				continue
			}
			mi, ok := ev.MatMap[ms]
			if !ok {
				fmt.Printf("Mat not found: %s\n", ms)
			} else {
				ev.World.Set([]int{y, x}, mi)
			}
		}
	}
	return nil
}

// RenderDepthView renders the depth view from current point, angle
func (ev *FWorld) RenderDepthView() {
	// ray-trace along vector bascially
}

// Step is called to advance the environment state
func (ev *FWorld) Step() bool {
	ev.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	// ev.NewPoint()
	if ev.Trial.Incr() { // true if wraps around Max back to 0
		ev.Epoch.Incr()
	}
	return true
}

func (ev *FWorld) Action(element string, input etensor.Tensor) {
	// nop
}

func (ev *FWorld) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return ev.Run.Query()
	case env.Epoch:
		return ev.Epoch.Query()
	case env.Trial:
		return ev.Trial.Query()
	case env.Event:
		return ev.Event.Query()
	case env.Scene:
		return ev.Scene.Query()
	case env.Episode:
		return ev.Episode.Query()
	}
	return -1, -1, false
}

// Compile-time check that implements Env interface
var _ env.Env = (*FWorld)(nil)
