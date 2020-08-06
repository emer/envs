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
	Saccade Saccade           `desc:"the env item"`
	View    *etview.TableView `view:"-" desc:"the main view"`
	Win     *gi.Window        `view:"-" desc:"main GUI window"`
	ToolBar *gi.ToolBar       `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.Saccade.Defaults()
	// ss.Saccade.AddRows = true
	ss.Saccade.Init()
	ss.Saccade.Step() // need one in there
}

// Equation for biexponential synapse from here:
// https://brian2.readthedocs.io/en/stable/user/converting_from_integrated_form.html

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	// gi.WinEventTrace = true

	gi.SetAppName("env")
	gi.SetAppAbout(`This tests and Env. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)

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
	ss.View.SetTable(ss.Saccade.Table, nil)

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Saccade.Init()
		ss.View.SetTable(ss.Saccade.Table, nil)
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step", Tooltip: "Step env."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Saccade.Step()
		ss.View.SetTable(ss.Saccade.Table, nil)
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
