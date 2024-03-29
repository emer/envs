// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/evec"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// XYHDEnv is a flat-world grid-based environment with XY position and Head Direction, adapted from fworld
type XYHDEnv struct {

	// name of this environment
	Nm string `desc:"name of this environment"`

	// description of this environment
	Dsc string `desc:"description of this environment"`

	// update display -- turn off to make it faster
	Disp bool `desc:"update display -- turn off to make it faster"`

	// size of 2D world
	Size evec.Vec2i `desc:"size of 2D world"`

	// size of patterns for mats, acts
	PatSize evec.Vec2i `desc:"size of patterns for mats, acts"`

	// size of patterns for xy coordinates
	PosSize evec.Vec2i `desc:"size of patterns for xy coordinates"`

	// [view: no-inline] 2D grid world, each cell is a material (mat)
	World *etensor.Int `view:"no-inline" desc:"2D grid world, each cell is a material (mat)"`

	// list of materials in the world, 0 = empty.  Any superpositions of states need to be discretely encoded, can be transformed through action rules
	Mats []string `desc:"list of materials in the world, 0 = empty.  Any superpositions of states need to be discretely encoded, can be transformed through action rules"`

	// map of material name to index stored in world cell
	MatMap map[string]int `desc:"map of material name to index stored in world cell"`

	// index of material below which (inclusive) cannot move -- e.g., 1 for wall
	BarrierIdx int `desc:"index of material below which (inclusive) cannot move -- e.g., 1 for wall"`

	// patterns for each material (must include Empty) and for each action
	Pats map[string]*etensor.Float32 `desc:"patterns for each material (must include Empty) and for each action"`

	// list of actions: starts with: Left, Right, Forward
	Acts []string `desc:"list of actions: starts with: Left, Right, Forward"`

	// action map of action names to indexes
	ActMap map[string]int `desc:"action map of action names to indexes"`

	// map of optional interoceptive and world-dynamic parameters -- cleaner to store in a map
	Params map[string]float32 `desc:"map of optional interoceptive and world-dynamic parameters -- cleaner to store in a map"`

	// angle increment for rotation, in degrees -- defaults to 90
	AngInc int `desc:"angle increment for rotation, in degrees -- defaults to 90"`

	// total number of rotation angles in a circle
	NRotAngles int `inactive:"+" desc:"total number of rotation angles in a circle"`

	// for debugging, print out a trace of the action generation logic
	TraceActGen bool `desc:"for debugging, print out a trace of the action generation logic"`

	// number of units in ring population codes
	RingSize int `inactive:"+" desc:"number of units in ring population codes"`

	// number of units in population codes
	VesSize int `inactive:"+" desc:"number of units in population codes"`

	// population code values, in normalized units
	PopCode popcode.OneD `desc:"population code values, in normalized units"`

	// 2d population code values, in normalized units
	PopCode2d popcode.TwoD `desc:"2d population code values, in normalized units"`

	// angle population code values, in normalized units
	AngCode popcode.Ring `desc:"angle population code values, in normalized units"`

	// current location of agent, floating point
	PrevPosF mat32.Vec2 `inactive:"+" desc:"current location of agent, floating point"`

	// current location of agent, integer
	PrevPosI evec.Vec2i `inactive:"+" desc:"current location of agent, integer"`

	// current location of agent, floating point
	PosF mat32.Vec2 `inactive:"+" desc:"current location of agent, floating point"`

	// current location of agent, integer
	PosI evec.Vec2i `inactive:"+" desc:"current location of agent, integer"`

	// current angle, in degrees
	PrevAngle int `inactive:"+" desc:"current angle, in degrees"`

	// current angle, in degrees
	Angle int `inactive:"+" desc:"current angle, in degrees"`

	// angle that we just rotated -- drives vestibular
	RotAng int `inactive:"+" desc:"angle that we just rotated -- drives vestibular"`

	// last action taken
	Act int `inactive:"+" desc:"last action taken"`

	// material at each right angle: front, right, left, back
	ProxMats []int `desc:"material at each right angle: front, right, left, back"`

	// coordinates for proximal grid points: front, right, left, back
	ProxPos []evec.Vec2i `desc:"coordinates for proximal grid points: front, right, left, back"`

	// current rendered state tensors -- extensible map
	CurStates map[string]*etensor.Float32 `desc:"current rendered state tensors -- extensible map"`

	// next rendered state tensors -- updated from actions
	NextStates map[string]*etensor.Float32 `desc:"next rendered state tensors -- updated from actions"`

	// list of events, key is tick step, to check each step to drive refresh of consumables -- removed from this active list when complete
	RefreshEvents map[int]*WEvent `desc:"list of events, key is tick step, to check each step to drive refresh of consumables -- removed from this active list when complete"`

	// list of all events, key is tick step
	AllEvents map[int]*WEvent `desc:"list of all events, key is tick step"`

	// [view: inline] current run of model as provided during Init
	Run env.Ctr `view:"inline" desc:"current run of model as provided during Init"`

	// [view: inline] increments over arbitrary fixed number of trials, for general stats-tracking
	Epoch env.Ctr `view:"inline" desc:"increments over arbitrary fixed number of trials, for general stats-tracking"`

	// [view: inline] increments for each step of world, loops over epochs -- for general stats-tracking independent of env state
	Trial env.Ctr `view:"inline" desc:"increments for each step of world, loops over epochs -- for general stats-tracking independent of env state"`

	// [view: monolithic time counter -- counts up time every step -- used for refreshing world state]
	Tick env.Ctr `view:"monolithic time counter -- counts up time every step -- used for refreshing world state"`

	// [view: arbitrary counter for steps within a scene -- resets at consumption event]
	Event env.Ctr `view:"arbitrary counter for steps within a scene -- resets at consumption event"`

	// [view: arbitrary counter incrementing over a coherent sequence of events: e.g., approaching food -- increments at consumption]
	Scene env.Ctr `view:"arbitrary counter incrementing over a coherent sequence of events: e.g., approaching food -- increments at consumption"`

	// [view: arbitrary counter incrementing over scenes within larger episode: feeding, drinking, exploring, etc]
	Episode env.Ctr `view:"arbitrary counter incrementing over scenes within larger episode: feeding, drinking, exploring, etc"`
}

var KiT_XYHDEnv = kit.Types.AddType(&XYHDEnv{}, XYHDEnvProps)

func (ev *XYHDEnv) Name() string { return ev.Nm }
func (ev *XYHDEnv) Desc() string { return ev.Dsc }

// Config configures the world
func (ev *XYHDEnv) Config(ntrls int) {
	ev.Nm = "Demo"
	ev.Dsc = "Example world with xy coordinate system and head direction"
	ev.Mats = []string{"Empty", "Wall"}
	ev.BarrierIdx = 1
	ev.Acts = []string{"Left", "Right", "Forward"}
	ev.Params = make(map[string]float32)

	ev.Disp = false
	ev.Size.Set(200, 200) // if changing to non-square, reset the popcode2d min
	ev.PatSize.Set(5, 5)
	ev.PosSize.Set(12, 12)
	ev.AngInc = 90
	ev.RingSize = 16 // was 16
	ev.VesSize = 12  // was 12
	ev.PopCode.Defaults()
	ev.PopCode.SetRange(-0.2, 1.2, 0.1)
	ev.PopCode2d.Defaults()
	ev.PopCode2d.SetRange(1/(float32(ev.Size.X)-2), 1, 0.2) // assume it's a square, 2 is length of walls
	//ev.PopCode2d.SetRange(0, 1, 0.1) // assume it's a square, 2 is length of walls
	ev.AngCode.Defaults()
	ev.AngCode.SetRange(0, 1, 0.1) // zycyc experiment

	// debugging options:
	ev.TraceActGen = false

	ev.Trial.Max = ntrls

	ev.ConfigPats()
	ev.ConfigImpl()

	// uncomment to generate a new world
	ev.GenWorld()
	//ev.SaveWorld("world.tsv")
}

// ConfigPats configures the bit pattern representations of mats and acts
func (ev *XYHDEnv) ConfigPats() {
	ev.Pats = make(map[string]*etensor.Float32)
	for _, m := range ev.Mats {
		t := &etensor.Float32{}
		t.SetShape([]int{ev.PatSize.Y, ev.PatSize.X}, nil, []string{"Y", "X"})
		ev.Pats[m] = t
	}
	for _, a := range ev.Acts {
		t := &etensor.Float32{}
		t.SetShape([]int{ev.PatSize.Y, ev.PatSize.X}, nil, []string{"Y", "X"})
		ev.Pats[a] = t
	}
	ev.OpenPats("pats.json") // hand crafted..
}

// ConfigImpl does the automatic parts of configuration
// generally does not require editing
func (ev *XYHDEnv) ConfigImpl() {
	ev.NRotAngles = (360 / ev.AngInc) + 1

	ev.World = &etensor.Int{}
	ev.World.SetShape([]int{ev.Size.Y, ev.Size.X}, nil, []string{"Y", "X"})

	ev.ProxMats = make([]int, 4)
	ev.ProxPos = make([]evec.Vec2i, 4)

	ev.CurStates = make(map[string]*etensor.Float32)
	ev.NextStates = make(map[string]*etensor.Float32)

	ps := &etensor.Float32{}
	ps.SetShape([]int{1, 4, 2, 1}, nil, []string{"1", "Pos", "OnOff", "1"})
	ev.NextStates["ProxSoma"] = ps

	ag := &etensor.Float32{}
	ag.SetShape([]int{1, ev.RingSize}, nil, []string{"1", "Pop"})
	ev.NextStates["Angle"] = ag

	prevag := &etensor.Float32{}
	prevag.SetShape([]int{1, ev.RingSize}, nil, []string{"1", "Pop"})
	ev.NextStates["PrevAngle"] = prevag

	vs := &etensor.Float32{}
	vs.SetShape([]int{1, ev.VesSize}, nil, []string{"1", "Pop"})
	ev.NextStates["Vestibular"] = vs

	xy := &etensor.Float32{}
	xy.SetShape([]int{ev.PosSize.Y, ev.PosSize.X}, nil, []string{"Y", "X"})
	ev.NextStates["Position"] = xy

	prevxy := &etensor.Float32{}
	prevxy.SetShape([]int{ev.PosSize.Y, ev.PosSize.X}, nil, []string{"Y", "X"})
	ev.NextStates["PrevPosition"] = prevxy

	av := &etensor.Float32{}
	av.SetShape([]int{ev.PatSize.Y, ev.PatSize.X}, nil, []string{"Y", "X"})
	ev.NextStates["Action"] = av

	ev.CopyNextToCur() // get CurStates from NextStates

	ev.MatMap = make(map[string]int, len(ev.Mats))
	for i, m := range ev.Mats {
		ev.MatMap[m] = i
	}
	ev.ActMap = make(map[string]int, len(ev.Acts))
	for i, m := range ev.Acts {
		ev.ActMap[m] = i
	}

	ev.Run.Scale = env.Run
	ev.Epoch.Scale = env.Epoch
	ev.Trial.Scale = env.Trial
	ev.Tick.Scale = env.Tick
	ev.Event.Scale = env.Event
	ev.Scene.Scale = env.Scene
	ev.Episode.Scale = env.Episode
}

func (ev *XYHDEnv) Validate() error {
	if ev.Size.IsNil() {
		return fmt.Errorf("XYHDEnv: %v has size == 0 -- need to Config", ev.Nm)
	}
	return nil
}

func (ev *XYHDEnv) State(element string) etensor.Tensor {
	return ev.CurStates[element]
}

// String returns the current state as a string
func (ev *XYHDEnv) String() string {
	return fmt.Sprintf("Evt_%d_Pos_%d_%d_Ang_%d_Act_%s", ev.Event.Cur, ev.PosI.X, ev.PosI.Y, ev.Angle, ev.Acts[ev.Act])
}

// Init is called to restart environment
func (ev *XYHDEnv) Init(run int) {

	// note: could gen a new random world too..
	//ev.OpenWorld("world.tsv")

	ev.Run.Init()
	ev.Epoch.Init()
	ev.Trial.Init()
	ev.Tick.Init()
	ev.Event.Init()
	ev.Scene.Init()
	ev.Episode.Init()

	ev.Run.Cur = run
	ev.Trial.Cur = -1 // init state -- key so that first Step() = 0
	ev.Tick.Cur = -1
	ev.Event.Cur = -1

	ev.PosI = ev.Size.DivScalar(2) // start in middle -- could be random..
	ev.PosF = ev.PosI.ToVec2()
	for i := 0; i < 4; i++ {
		ev.ProxMats[i] = 0
	}

	ev.Angle = 0
	ev.RotAng = 0

	ev.RefreshEvents = make(map[int]*WEvent)
	ev.AllEvents = make(map[int]*WEvent)
}

// SetWorld sets given mat at given point coord in world
func (ev *XYHDEnv) SetWorld(p evec.Vec2i, mat int) {
	ev.World.Set([]int{p.Y, p.X}, mat)
}

// GetWorld returns mat at given point coord in world
func (ev *XYHDEnv) GetWorld(p evec.Vec2i) int {
	return ev.World.Value([]int{p.Y, p.X})
}

////////////////////////////////////////////////////////////////////
// I/O

// SaveWorld saves the world to a tsv file with empty string for empty cells
func (ev *XYHDEnv) SaveWorld(filename gi.FileName) error {
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
func (ev *XYHDEnv) OpenWorld(filename gi.FileName) error {
	fp, err := os.Open(string(filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer fp.Close()
	ev.World.SetZeros()
	scan := bufio.NewScanner(fp)
	for y := 0; y < ev.Size.Y; y++ {
		if !scan.Scan() {
			break
		}
		ln := scan.Bytes()
		sz := len(ln)
		if sz == 0 {
			break
		}
		sp := bytes.Split(ln, []byte("\t"))
		sz = ints.MinInt(ev.Size.X, len(sp)-1)
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

// SavePats saves the patterns
func (ev *XYHDEnv) SavePats(filename gi.FileName) error {
	jenc, _ := json.MarshalIndent(ev.Pats, "", " ")
	return ioutil.WriteFile(string(filename), jenc, 0644)
}

// OpenPats opens the patterns
func (ev *XYHDEnv) OpenPats(filename gi.FileName) error {
	fp, err := os.Open(string(filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer fp.Close()
	b, err := ioutil.ReadAll(fp)
	err = json.Unmarshal(b, &ev.Pats)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// AngMod returns angle modulo within 360 degrees
func AngMod(ang int) int {
	if ang < 0 {
		ang += 360
	} else if ang > 360 {
		ang -= 360
	}
	return ang
}

// AngVec returns the incremental vector to use for given angle, in deg
// such that the largest value is 1.
func AngVec(ang int) mat32.Vec2 {
	a := mat32.DegToRad(float32(AngMod(ang)))
	v := mat32.Vec2{mat32.Cos(a), mat32.Sin(a)}
	return NormVecLine(v)
}

// NormVec normalize vector for drawing a line
func NormVecLine(v mat32.Vec2) mat32.Vec2 {
	av := v.Abs()
	if av.X > av.Y {
		v = v.DivScalar(av.X)
	} else {
		v = v.DivScalar(av.Y)
	}
	return v
}

// NextVecPoint returns the next grid point along vector,
// from given current floating and grid points.  v is normalized
// such that the largest value is 1.
func NextVecPoint(cp, v mat32.Vec2) (mat32.Vec2, evec.Vec2i) {
	n := cp.Add(v)
	g := evec.NewVec2iFmVec2Round(n)
	return n, g
}

////////////////////////////////////////////////////////////////////
// Vision

// ScanProx scan the proximal space around the agent
func (ev *XYHDEnv) ScanProx() {
	angs := []int{0, -90, 90, 180}
	for i := 0; i < 4; i++ {
		v := AngVec(ev.Angle + angs[i])
		_, gp := NextVecPoint(ev.PosF, v)
		ev.ProxMats[i] = ev.GetWorld(gp)
		ev.ProxPos[i] = gp
	}
}

////////////////////////////////////////////////////////////////////
// Actions

// WEvent records an event
type WEvent struct {

	// tick when event happened
	Tick int `desc:"tick when event happened"`

	// discrete integer grid position where event happened
	PosI evec.Vec2i `desc:"discrete integer grid position where event happened"`

	// floating point grid position where event happened
	PosF mat32.Vec2 `desc:"floating point grid position where event happened"`

	// angle pointing when event happened
	Angle int `desc:"angle pointing when event happened"`

	// action that took place
	Act int `desc:"action that took place"`

	// material that was involved (front fovea mat)
	Mat int `desc:"material that was involved (front fovea mat)"`

	// position of material involved in event
	MatPos evec.Vec2i `desc:"position of material involved in event"`
}

// NewEvent returns new event with current state and given act, mat
func (ev *XYHDEnv) NewEvent(act, mat int, matpos evec.Vec2i) *WEvent {
	return &WEvent{Tick: ev.Tick.Cur, PosI: ev.PosI, PosF: ev.PosF, Angle: ev.Angle, Act: act, Mat: mat, MatPos: matpos}
}

// AddNewEventRefresh adds event to RefreshEvents (a consumable was consumed).
// always adds to AllEvents
func (ev *XYHDEnv) AddNewEventRefresh(wev *WEvent) {
	ev.RefreshEvents[wev.Tick] = wev
	ev.AllEvents[wev.Tick] = wev
}

// TakeAct takes the action, updates state
func (ev *XYHDEnv) TakeAct(act int) {
	//as := ""
	//if act >= len(ev.Acts) || act < 0 {
	//	as = "Stay"
	//} else {
	//	as = ev.Acts[act]
	//}

	as := ev.Acts[act]
	ev.RotAng = 0

	nmat := len(ev.Mats)
	frmat := ints.MinInt(ev.ProxMats[0], nmat)
	//behmat := ev.ProxMats[3] // behind

	ev.PrevPosF, ev.PrevPosI = ev.PosF, ev.PosI
	ev.PrevAngle = ev.Angle
	switch as {
	//case "Stay":
	case "Left":
		ev.RotAng = ev.AngInc
		ev.Angle = AngMod(ev.Angle + ev.RotAng)
		ev.PosF, ev.PosI = NextVecPoint(ev.PosF, AngVec(ev.Angle)) // when L/R contains forward
	case "Right":
		ev.RotAng = -ev.AngInc
		ev.Angle = AngMod(ev.Angle + ev.RotAng)
		ev.PosF, ev.PosI = NextVecPoint(ev.PosF, AngVec(ev.Angle)) // when L/R contains forward
	case "Forward":
		if frmat > 0 && frmat <= ev.BarrierIdx {
		} else {
			ev.PosF, ev.PosI = NextVecPoint(ev.PosF, AngVec(ev.Angle))
		}
		//case "Backward":
		//	if behmat > 0 && behmat <= ev.BarrierIdx {
		//	} else {
		//		ev.PosF, ev.PosI = NextVecPoint(ev.PosF, AngVec(AngMod(ev.Angle+180)))
		//	}
	}
	ev.ScanProx()

	ev.RenderState()
}

// RenderProxSoma renders proximal soma state
func (ev *XYHDEnv) RenderProxSoma() {
	ps := ev.NextStates["ProxSoma"]
	ps.SetZeros()
	for i := 0; i < 4; i++ {
		if ev.ProxMats[i] != 0 {
			ps.Set([]int{0, i, 0, 0}, 1) // on
		} else {
			ps.Set([]int{0, i, 1, 0}, 1) // off
		}
	}
}

// RenderAngle renders angle using pop ring
func (ev *XYHDEnv) RenderAngle(statenm string, angle int) {
	as := ev.NextStates[statenm]
	av := (float32(angle) / 360.0)
	ev.AngCode.Encode(&as.Values, av, ev.RingSize)

	//as.SetZeros()
	//if angle == 0 || angle == 360 {
	//	as.Values = []float32{0, 1, 0, 1}
	//} else if angle == 90 {
	//	as.Values = []float32{0, 0, 1, 1}
	//} else if angle == 180 {
	//	as.Values = []float32{1, 0, 1, 0}
	//} else if angle == 270 {
	//	as.Values = []float32{1, 1, 0, 0}
	//}

}

// RenderVestib renders vestibular state
func (ev *XYHDEnv) RenderVestibular() {
	vs := ev.NextStates["Vestibular"]
	nv := 0.5*(float32(-ev.RotAng)/90) + 0.5
	ev.PopCode.Encode(&vs.Values, nv, ev.VesSize, false)

	//vs.SetZeros()
	//if ev.RotAng == -90 {
	//	vs.Values = []float32{1, 0, 0}
	//} else if ev.RotAng == 0 {
	//	vs.Values = []float32{0, 1, 0}
	//} else if ev.RotAng == 90 {
	//	vs.Values = []float32{0, 0, 1}
	//}

}

// RenderPosition renders position using 2d popcode
func (ev *XYHDEnv) RenderPosition(statenm string, posf mat32.Vec2) {
	xy := ev.NextStates[statenm]
	pv := posf
	pv.X /= float32(ev.Size.X) - 2
	pv.Y /= float32(ev.Size.Y) - 2
	ev.PopCode2d.Encode(xy, pv, false)
}

// RenderAction renders action pattern
func (ev *XYHDEnv) RenderAction() {
	av := ev.NextStates["Action"]
	if ev.Act < len(ev.Acts) {
		as := ev.Acts[ev.Act]
		ap, ok := ev.Pats[as]
		if ok {
			av.CopyFrom(ap)
		}
	}
}

// RenderState renders the current state into NextState vars
func (ev *XYHDEnv) RenderState() {
	ev.RenderProxSoma()
	ev.RenderAngle("Angle", ev.Angle)
	ev.RenderAngle("PrevAngle", ev.PrevAngle)
	ev.RenderVestibular()
	ev.RenderPosition("Position", ev.PosF)
	ev.RenderPosition("PrevPosition", ev.PrevPosF)
	ev.RenderAction()
}

// CopyNextToCur copy next state to current state
func (ev *XYHDEnv) CopyNextToCur() {
	for k, ns := range ev.NextStates {
		cs, ok := ev.CurStates[k]
		if !ok {
			cs = ns.Clone().(*etensor.Float32)
			ev.CurStates[k] = cs
		} else {
			cs.CopyFrom(ns)
		}
	}
}

// Step is called to advance the environment state
func (ev *XYHDEnv) Step() bool {
	ev.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	ev.CopyNextToCur()
	ev.Tick.Incr()
	ev.Event.Incr()
	if ev.Trial.Incr() { // true if wraps around Max back to 0
		ev.Epoch.Incr()
	}
	return true
}

func (ev *XYHDEnv) Action(action string, nop etensor.Tensor) {
	a, ok := ev.ActMap[action]
	if !ok {
		fmt.Printf("Action not recognized: %s\n", action)
		return
	}
	ev.Act = a
	ev.TakeAct(ev.Act)
}

func (ev *XYHDEnv) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return ev.Run.Query()
	case env.Epoch:
		return ev.Epoch.Query()
	case env.Trial:
		return ev.Trial.Query()
	case env.Tick:
		return ev.Tick.Query()
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
var _ env.Env = (*XYHDEnv)(nil)

var XYHDEnvProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"OpenWorld", ki.Props{
			"label": "Open World...",
			"icon":  "file-open",
			"desc":  "Open World from tsv file",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".tsv",
				}},
			},
		}},
		{"SaveWorld", ki.Props{
			"label": "Save World...",
			"icon":  "file-save",
			"desc":  "Save World to tsv file",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".tsv",
				}},
			},
		}},
		{"OpenPats", ki.Props{
			"label": "Open Pats...",
			"icon":  "file-open",
			"desc":  "Open pats from json file",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
		{"SavePats", ki.Props{
			"label": "Save Pats...",
			"icon":  "file-save",
			"desc":  "Save pats to json file",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
	},
}

////////////////////////////////////////////////////////////////////
// Render world

// WorldLineHoriz draw horizontal line
func (ev *XYHDEnv) WorldLineHoriz(st, ed evec.Vec2i, mat int) {
	sx := ints.MinInt(st.X, ed.X)
	ex := ints.MaxInt(st.X, ed.X)
	for x := sx; x <= ex; x++ {
		ev.World.Set([]int{st.Y, x}, mat)
	}
}

// WorldLineVert draw vertical line
func (ev *XYHDEnv) WorldLineVert(st, ed evec.Vec2i, mat int) {
	sy := ints.MinInt(st.Y, ed.Y)
	ey := ints.MaxInt(st.Y, ed.Y)
	for y := sy; y <= ey; y++ {
		ev.World.Set([]int{y, st.X}, mat)
	}
}

// WorldLine draw line in world with given mat
func (ev *XYHDEnv) WorldLine(st, ed evec.Vec2i, mat int) {
	di := ed.Sub(st)

	if di.X == 0 {
		ev.WorldLineVert(st, ed, mat)
		return
	}
	if di.Y == 0 {
		ev.WorldLineHoriz(st, ed, mat)
		return
	}

	dv := di.ToVec2()
	dst := dv.Length()
	v := NormVecLine(dv)
	op := st.ToVec2()
	cp := op
	gp := evec.Vec2i{}
	for {
		cp, gp = NextVecPoint(cp, v)
		ev.SetWorld(gp, mat)
		d := cp.DistTo(op) // not very efficient, but works.
		if d >= dst {
			break
		}
	}
}

// WorldRandom distributes n of given material in random locations
func (ev *XYHDEnv) WorldRandom(n, mat int) {
	cnt := 0
	for cnt < n {
		px := rand.Intn(ev.Size.X)
		py := rand.Intn(ev.Size.Y)
		ix := []int{py, px}
		cm := ev.World.Value(ix)
		if cm == 0 {
			ev.World.Set(ix, mat)
			cnt++
		}
	}
}

// WorldRect draw rectangle in world with given mat
func (ev *XYHDEnv) WorldRect(st, ed evec.Vec2i, mat int) {
	ev.WorldLineHoriz(st, evec.Vec2i{ed.X, st.Y}, mat)
	ev.WorldLineHoriz(evec.Vec2i{st.X, ed.Y}, evec.Vec2i{ed.X, ed.Y}, mat)
	ev.WorldLineVert(st, evec.Vec2i{st.X, ed.Y}, mat)
	ev.WorldLineVert(evec.Vec2i{ed.X, st.Y}, evec.Vec2i{ed.X, ed.Y}, mat)
}

// GenWorld generates a world -- edit to create in way desired
func (ev *XYHDEnv) GenWorld() {
	wall := ev.MatMap["Wall"]
	ev.World.SetZeros()
	// always start with a wall around the entire world -- no seeing the turtles..
	ev.WorldRect(evec.Vec2i{0, 0}, evec.Vec2i{ev.Size.X - 1, ev.Size.Y - 1}, wall)
	//ev.WorldRect(evec.Vec2i{20, 20}, evec.Vec2i{40, 40}, wall)
	//ev.WorldRect(evec.Vec2i{60, 60}, evec.Vec2i{80, 80}, wall)
	//
	//ev.WorldLine(evec.Vec2i{60, 20}, evec.Vec2i{80, 40}, wall) // double-thick lines = no leak
	//ev.WorldLine(evec.Vec2i{60, 19}, evec.Vec2i{80, 39}, wall)

	// don't put anything in center starting point
	ctr := ev.Size.DivScalar(2)
	ev.SetWorld(ctr, wall)

	// clear center
	ev.SetWorld(ctr, 0)
}

////////////////////////////////////////////////////////////////////
// Subcortex / Instinct

// ActGenTrace prints trace of act gen if enabled
func (ev *XYHDEnv) ActGenTrace(desc string, act int) {
	if !ev.TraceActGen {
		return
	}
	fmt.Printf("%s: act: %s\n", desc, ev.Acts[act])
}

// ActGen generates an action for current situation based on simple
// coded heuristics -- i.e., what subcortical evolutionary instincts provide.
func (ev *XYHDEnv) ActGen() int {
	wall := ev.MatMap["Wall"]
	left := ev.ActMap["Left"]
	right := ev.ActMap["Right"]

	nmat := len(ev.Mats)
	frmat := ints.MinInt(ev.ProxMats[0], nmat)
	rmat := ints.MinInt(ev.ProxMats[1], nmat)
	lmat := ints.MinInt(ev.ProxMats[2], nmat)

	rlp := float64(.5)
	rlact := left
	if erand.BoolP(rlp, -1) {
		rlact = right
	}
	rlps := fmt.Sprintf("%.3g", rlp)

	lastact := ev.Act
	frnd := rand.Float32()

	act := ev.ActMap["Forward"] // default

	// when L/R contains forward
	switch {
	case frmat == wall:
		if (rmat != wall) && (lmat != wall) {
			if lastact == left || lastact == right {
				act = lastact // keep going
				ev.ActGenTrace("at wall, keep turning", act)
			} else {
				act = rlact
				ev.ActGenTrace(fmt.Sprintf("at wall, rlp: %s, turn", rlps), act)
			}
		} else if rmat == wall {
			act = left
		} else {
			act = right
		}
	default: // random explore -- nothing obvious
		switch {
		//case frnd < 0.25:
		//	act = lastact // continue
		//	ev.ActGenTrace("repeat last act", act)
		case frnd < 0.15:
			if lmat == wall {
				act = right
			} else {
				act = left
			}
			ev.ActGenTrace("turn", act)
		case frnd < 0.3:
			if rmat == wall {
				act = left
			} else {
				act = right
			}
			ev.ActGenTrace("turn", act)
		default:
			ev.ActGenTrace("go", act)
		}
	}

	// when L/R doesn't contain forward
	//switch {
	//case frmat == wall:
	//	if lastact == left || lastact == right {
	//		act = lastact // keep going
	//		ev.ActGenTrace("at wall, keep turning", act)
	//	} else {
	//		act = rlact
	//		ev.ActGenTrace(fmt.Sprintf("at wall, rlp: %s, turn", rlps), act)
	//	}
	//default: // random explore -- nothing obvious
	//	switch {
	//	//case frnd < 0.25:
	//	//	act = lastact // continue
	//	//	ev.ActGenTrace("repeat last act", act)
	//	case frnd < 0.15:
	//		act = left
	//		ev.ActGenTrace("turn", act)
	//	case frnd < 0.3:
	//		act = right
	//		ev.ActGenTrace("turn", act)
	//	default:
	//		ev.ActGenTrace("go", act)
	//	}
	//}

	return act
}
