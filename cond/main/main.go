// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
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
	Env       Approach    `desc:"the env item"`
	Grids     *gi.Layout  `view:"-" desc:"the grids"`
	GridNames []string    `view:"-" desc:"state names for each grid"`
	Win       *gi.Window  `view:"-" desc:"main GUI window"`
	ToolBar   *gi.ToolBar `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.Env.Defaults()
	// ss.Sac.AddRows = true
	ss.Env.Init(0)
	ss.Env.Step() // need one in there
}

func (ss *Sim) UpdateGrids() {
	for i := range ss.GridNames {
		tg := ss.Grids.Child(i*3 + 1).(*etview.TensorGrid)
		tg.UpdateSig()
	}
}

func (ss *Sim) Step() {
	act := ss.Env.ActGen()
	actNm := ss.Env.Acts[act]
	ss.Env.Action(actNm, nil)
	ss.Env.Step()
	// ss.View.UpdateTable()

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
	sv.SetStruct(&ss.Env)

	tv := gi.AddNewTabView(split, "tv")
	ss.Grids = tv.AddNewTab(gi.KiT_Layout, "Grids").(*gi.Layout)
	ss.Grids.Lay = gi.LayoutVert

	ss.GridNames = []string{"USs", "CSs", "Pos", "Drives", "US", "Rew", "CS", "Dist", "Time", "Action"}
	for _, gr := range ss.GridNames {
		tg := &etview.TensorGrid{}
		tg.SetName(gr)
		gi.AddNewLabel(ss.Grids, gr, gr+":")
		ss.Grids.AddChild(tg)
		tg.SetTensor(ss.Env.States[gr])
		if gr != "Rew" {
			tg.Disp.Range.FixMax = false
		}
		gi.AddNewSpace(ss.Grids, gr+"_spc")
		tg.SetStretchMax()
	}

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
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

	tbar.AddSeparator("run-sep")

	tbar.AddAction(gi.ActOpts{Label: "Forward", Icon: "wedge-up", Tooltip: "Step Forward"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Env.Action("Forward", nil)
		ss.UpdateGrids()
		vp.SetFullReRender()
		vp.UpdateSig()
	})

	tbar.AddAction(gi.ActOpts{Label: "Left", Icon: "wedge-left", Tooltip: "Rotate Left"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Env.Action("Left", nil)
		ss.UpdateGrids()
		vp.SetFullReRender()
		vp.UpdateSig()
	})

	tbar.AddAction(gi.ActOpts{Label: "Right", Icon: "wedge-right", Tooltip: "Rotate Right"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Env.Action("Right", nil)
		ss.UpdateGrids()
		vp.SetFullReRender()
		vp.UpdateSig()
	})

	tbar.AddAction(gi.ActOpts{Label: "Consume", Icon: "field", Tooltip: "Consume US -- only if zero distance"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Env.Action("Consume", nil)
		ss.UpdateGrids()
		vp.SetFullReRender()
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
