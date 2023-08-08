// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/etview"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
)

func main() {
	TheSim.Config()
	gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
		guirun()
	})
}

func guirun() {
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {

	// size of input
	InputSize evec.Vec2i `desc:"size of input"`

	// the env item
	Corpus CorpusEnv `desc:"the env item"`

	// table recording env
	Table *etable.Table `desc:"table recording env"`

	// [view: -] the main view
	View *etview.TableView `view:"-" desc:"the main view"`

	// [view: -] main GUI window
	Win *gi.Window `view:"-" desc:"main GUI window"`

	// [view: -] the master toolbar
	ToolBar *gi.ToolBar `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.InputSize = evec.Vec2i{25, 25}
	ss.Corpus.Nm = "Corpus"
	ss.Corpus.Dsc = "training params and state"
	ss.Corpus.Trial.Max = 1000
	ss.Corpus.Run.Max = 1

	ss.Corpus.Init(0)
	ss.Corpus.Config(false, "cbt_valid_filt.json", 5, ss.InputSize, false, false, 0.1)

	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT64, []int{ss.InputSize.Y, ss.InputSize.X}, nil},
	}
	ss.Table = etable.NewTable("input")
	ss.Table.SetFromSchema(sch, 1)
	ss.Table.SetMetaData("TrialName:width", "60")

	ss.Step() // need one in there
}

// Step takes one step and records in table
func (ss *Sim) Step() {
	ss.Corpus.Step()
	inp := ss.Corpus.State("Input")
	ss.Table.SetCellTensor("Input", 0, inp)
	ss.Table.SetCellString("TrialName", 0, ss.Corpus.String())
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	// gi.WinEventTrace = true

	gi.SetAppName("env")
	gi.SetAppAbout(`This tests an Env. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)

	win := gi.NewMainWindow("env", "Env Test", width, height)
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
	sv.SetStruct(ss)

	tv := gi.AddNewTabView(split, "tv")

	ss.View = tv.AddNewTab(etview.KiT_TableView, "Table").(*etview.TableView)
	ss.View.SetTable(ss.Table, nil)

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Corpus.Init(0)
		ss.View.UpdateTable()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Step()
		ss.View.UpdateTable()
		vp.SetNeedsFullRender()
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
