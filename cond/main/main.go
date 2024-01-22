// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
	"github.com/emer/envs/cond"
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
	win := TheSim.ConfigGUI()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {

	// run name
	RunName string `desc:"run name"`

	// the env item
	Env cond.CondEnv `desc:"the env item"`

	// [view: -] the grids
	Grids *gi.Layout `view:"-" desc:"the grids"`

	// [view: -] state names for each grid
	GridNames []string `view:"-" desc:"state names for each grid"`

	// [view: -] main GUI window
	Win *gi.Window `view:"-" desc:"main GUI window"`

	// [view: -] the master toolbar
	ToolBar *gi.ToolBar `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.RunName = "PosAcq_B50"
	ss.Env.Config(1, ss.RunName)
	// ss.Sac.AddRows = true
	ss.Env.Init(0)
}

func (ss *Sim) UpdateGrids() {
	for i := range ss.GridNames {
		tg := ss.Grids.Child(i*2 + 1).(*etview.TensorGrid)
		tg.UpdateSig()
	}
}

func (ss *Sim) Step() {
	ss.Env.Step()
	// ss.View.UpdateTable()
}

// ConfigGUI configures the Cogent Core gui interface for this simulation,
func (ss *Sim) ConfigGUI() *gi.Window {
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
	sv.SetStruct(&ss.Env)

	tv := gi.AddNewTabView(split, "tv")
	ss.Grids = tv.AddNewTab(gi.KiT_Layout, "Grids").(*gi.Layout)
	ss.Grids.Lay = gi.LayoutGrid
	ss.Grids.SetPropInt("columns", 4)
	ss.Grids.SetStretchMax()
	ss.GridNames = []string{"USpos", "USneg", "CS", "ContextIn", "USTimeIn"}
	for _, gr := range ss.GridNames {
		tg := &etview.TensorGrid{}
		tg.SetName(gr)
		gi.AddNewLabel(ss.Grids, gr, gr+":")
		ss.Grids.AddChild(tg)
		tg.SetTensor(ss.Env.State(gr))
		tg.Disp.Range.FixMax = true
		// gi.AddNewSpace(ss.Grids, gr+"_spc")
		tg.SetStretchMax()
	}

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Env.RunName = ss.RunName
		ss.Env.Init(0)
		// ss.View.UpdateTable()
		vp.SetNeedsFullRender()
		vp.UpdateSig()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Step()
		ss.UpdateGrids()
		vp.SetNeedsFullRender()
		vp.UpdateSig()
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
