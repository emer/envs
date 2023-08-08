// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/etview"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/goki/gi/colormap"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
)

func main() {
	gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
		guirun()
	})
}

func guirun() {
	TheSim.Config() // important to have this after gui init
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {

	// the flat world
	World FWorld `desc:"the flat world"`

	// number of steps to take for StepN button
	StepN int `desc:"number of steps to take for StepN button"`

	// [view: no-inline] trace of movement
	Trace *etensor.Int `view:"no-inline" desc:"trace of movement"`

	// view of the activity trace
	TraceView *etview.TensorGrid `desc:"view of the activity trace"`

	// view of the world
	WorldView *etview.TensorGrid `desc:"view of the world"`

	// table recording env
	State *etable.Table `desc:"table recording env"`

	// [view: -] the main view
	StateView *etview.TableView `view:"-" desc:"the main view"`

	// [view: -] main GUI window
	Win *gi.Window `view:"-" desc:"main GUI window"`

	// [view: -] the master toolbar
	ToolBar *gi.ToolBar `view:"-" desc:"the master toolbar"`

	// [view: -] the tab view
	TabView *gi.TabView `view:"-" desc:"the tab view"`

	// color strings in material order
	MatColors []string `desc:"color strings in material order"`

	// [view: -] flag to stop running
	StopNow bool `view:"-" desc:"flag to stop running"`

	// [view: -] true when running
	IsRunning bool `view:"-" desc:"true when running"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	// order: Empty, wall, food, water
	ss.MatColors = []string{"lightgrey", "black", "orange", "blue", "brown", "navy"}

	ss.StepN = 10
	ss.World.Config(1000)
	ss.World.Init(0)

	ss.Trace = ss.World.World.Clone().(*etensor.Int)

	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"Depth", etensor.FLOAT32, ss.World.CurStates["Depth"].Shape.Shp, nil},
		{"FovDepth", etensor.FLOAT32, ss.World.CurStates["FovDepth"].Shape.Shp, nil},
		{"Fovea", etensor.FLOAT32, ss.World.CurStates["Fovea"].Shape.Shp, nil},
		{"ProxSoma", etensor.FLOAT32, ss.World.CurStates["ProxSoma"].Shape.Shp, nil},
		{"Vestibular", etensor.FLOAT32, ss.World.CurStates["Vestibular"].Shape.Shp, nil},
		{"Inters", etensor.FLOAT32, ss.World.CurStates["Inters"].Shape.Shp, nil},
		{"Action", etensor.FLOAT32, ss.World.CurStates["Action"].Shape.Shp, nil},
	}
	ss.State = etable.NewTable("input")
	ss.State.SetFromSchema(sch, 1)
	ss.State.SetMetaData("TrialName:width", "50")
}

func (ss *Sim) UpdtViews() {
	updt := ss.TabView.UpdateStart()
	// ss.TraceView.UpdateSig()
	// ss.WorldView.UpdateSig()
	ss.StateView.UpdateTable()
	ss.TabView.UpdateEnd(updt)
}

// Step takes one step and records in table
func (ss *Sim) Step() {
	ss.World.Step()
	for i := 1; i < ss.State.NumCols(); i++ {
		cnm := ss.State.ColNames[i]
		inp := ss.World.State(cnm)
		ss.State.SetCellTensor(cnm, 0, inp)
	}
	ss.State.SetCellString("TrialName", 0, ss.World.String())

	if ss.World.Scene.Chg { // something important happened, refresh
		ss.Trace.CopyFrom(ss.World.World)
	}

	nc := len(ss.World.Mats)
	ss.Trace.Set([]int{ss.World.PosI.Y, ss.World.PosI.X}, nc+ss.World.Angle/ss.World.AngInc)

	ss.UpdtViews()
}

func (ss *Sim) StepAuto() {
	act := ss.World.ActGen()
	ss.World.Action(ss.World.Acts[act], nil)
	ss.Step()
}

func (ss *Sim) StepAutoN() {
	for i := 0; i < ss.StepN; i++ {
		ss.StepAuto()
	}
}

func (ss *Sim) Run() {
	ss.IsRunning = true
	ss.ToolBar.UpdateActions()
	for !ss.StopNow {
		ss.Win.Viewport.SetFullReRender()
		ss.StepAuto()
		ss.Win.PollEvents() // needed when not running in parallel.
	}
	ss.IsRunning = false
	ss.StopNow = false
	ss.ToolBar.UpdateActions()
}

func (ss *Sim) Stop() {
	ss.StopNow = true
}

func (ss *Sim) Init() {
	ss.World.Init(0)
	ss.Trace.CopyFrom(ss.World.World)
	ss.UpdtViews()
}

func (ss *Sim) Left() {
	ss.World.Action("Left", nil)
	ss.Step()
}

func (ss *Sim) Right() {
	ss.World.Action("Right", nil)
	ss.Step()
}

func (ss *Sim) Forward() {
	ss.World.Action("Forward", nil)
	ss.Step()
}

func (ss *Sim) Backward() {
	ss.World.Action("Backward", nil)
	ss.Step()
}

func (ss *Sim) Eat() {
	ss.World.Action("Eat", nil)
	ss.Step()
}

func (ss *Sim) Drink() {
	ss.World.Action("Drink", nil)
	ss.Step()
}

func (ss *Sim) ConfigWorldView(tg *etview.TensorGrid) {
	cnm := "FWorldColors"
	cm, ok := colormap.AvailMaps[cnm]
	if !ok {
		cm = &colormap.Map{}
		cm.Name = cnm
		cm.Indexed = true
		nc := len(ss.World.Mats)
		cm.Colors = make([]gist.Color, nc+ss.World.NRotAngles)
		cm.NoColor = gist.Black
		for i, cnm := range ss.MatColors {
			cm.Colors[i].SetString(cnm, nil)
		}
		ch := colormap.AvailMaps["ColdHot"]
		for i := 0; i < ss.World.NRotAngles; i++ {
			nv := float64(i) / float64(ss.World.NRotAngles-1)
			cm.Colors[nc+i] = ch.Map(nv) // color map of rotation
		}
		colormap.AvailMaps[cnm] = cm
	}
	tg.Disp.Defaults()
	tg.Disp.ColorMap = giv.ColorMapName(cnm)
	tg.Disp.GridFill = 1
	tg.SetStretchMax()
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	// gi.WinEventTrace = true

	gi.SetAppName("fworld")
	gi.SetAppAbout(`This tests an Env. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)

	win := gi.NewMainWindow("fworld", "Flat World", width, height)
	ss.Win = win

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tbar := gi.AddNewToolBar(mfr, "tbar")
	tbar.SetStretchMaxWidth()
	ss.ToolBar = tbar

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = mat32.X
	split.SetStretchMax()

	sv := giv.AddNewStructView(split, "sv")
	sv.SetStruct(&ss.World)

	tv := gi.AddNewTabView(split, "tv")
	ss.TabView = tv

	sps := tv.AddNewTab(gi.KiT_SplitView, "State").(*gi.SplitView)
	sps.Dim = mat32.Y
	sps.SetStretchMax()

	ss.StateView = etview.AddNewTableView(sps, "State")
	ss.StateView.SetTable(ss.State, nil)

	ss.TraceView = etview.AddNewTensorGrid(sps, "Trace", ss.Trace)
	ss.ConfigWorldView(ss.TraceView)

	sps.SetSplits(.3, .7)

	wg := tv.AddNewTab(etview.KiT_TensorGrid, "World").(*etview.TensorGrid)
	ss.WorldView = wg
	wg.SetTensor(ss.World.World)
	ss.ConfigWorldView(wg)

	split.SetSplits(.3, .7)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Init()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step one auto-generated action", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.StepAuto()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "StepN", Icon: "forward", Tooltip: "Step N auto-generated actions", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.StepAutoN()
		vp.SetFullReRender()
	})

	tbar.AddSeparator("sep-step")

	tbar.AddAction(gi.ActOpts{Label: "Run", Icon: "play", Tooltip: "run until stop pressed", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Run() // go run is too crashy -- hangs or crashes.  re-enable to debug
		tbar.UpdateActions()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop", Tooltip: "stop running", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Stop()
		tbar.UpdateActions()
		vp.SetNeedsFullRender()
	})

	tbar.AddSeparator("run-sep")

	tbar.AddAction(gi.ActOpts{Label: "Left", Icon: "wedge-left", Tooltip: "Rotate Left", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Left()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Right", Icon: "wedge-right", Tooltip: "Rotate Right", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Right()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Forward", Icon: "wedge-up", Tooltip: "Step Forward", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Forward()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Backward", Icon: "wedge-down", Tooltip: "Step Backward", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Backward()
		vp.SetFullReRender()
	})

	tbar.AddSeparator("sep-eat")

	tbar.AddAction(gi.ActOpts{Label: "Eat", Icon: "field", Tooltip: "Eat food -- only if directly in front", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Eat()
		vp.SetFullReRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Drink", Icon: "svg", Tooltip: "Drink water -- only if directly in front", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Drink()
		vp.SetFullReRender()
	})

	tbar.AddSeparator("sep-file")

	tbar.AddAction(gi.ActOpts{Label: "Open World", Icon: "file-open", Tooltip: "Open World from .tsv file", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		giv.CallMethod(&ss.World, "OpenWorld", vp)
	})

	tbar.AddAction(gi.ActOpts{Label: "Save World", Icon: "file-save", Tooltip: "Save World to .tsv file", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		giv.CallMethod(&ss.World, "SaveWorld", vp)
	})

	tbar.AddAction(gi.ActOpts{Label: "Open Pats", Icon: "file-open", Tooltip: "Open bit patterns from .json file", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		giv.CallMethod(&ss.World, "OpenPats", vp)
	})

	tbar.AddAction(gi.ActOpts{Label: "Save Pats", Icon: "file-save", Tooltip: "Save bit patterns to .json file", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		giv.CallMethod(&ss.World, "SavePats", vp)
	})

	vp.UpdateEndNoSig(updt)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	win.MainMenuUpdated()
	return win
}
