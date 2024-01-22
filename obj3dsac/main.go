// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
	"github.com/emer/etable/etview"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/gimain"
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
	win := TheSim.ConfigGUI()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {

	// the env item
	Obj Obj3DSac `desc:"the env item"`

	// number of steps to take for StepN button
	StepN int `desc:"number of steps to take for StepN button"`

	// [view: -] the main view
	TableView *etview.TableView `view:"-" desc:"the main view"`

	// [view: -] main GUI window
	Win *gi.Window `view:"-" desc:"main GUI window"`

	// [view: -] the master toolbar
	ToolBar *gi.ToolBar `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.StepN = 8
	ss.Obj.Defaults()
	ss.Obj.Config()
	ss.Obj.Init()
}

// ConfigGUI configures the Cogent Core gui interface for this simulation,
func (ss *Sim) ConfigGUI() *gi.Window {
	width := 1600
	height := 1200

	// gi.WinEventTrace = true

	gi.SetAppName("obj3dsac")
	gi.SetAppAbout(`This tests an Env. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)

	win := gi.NewMainWindow("obj3dsac", "Obj3D Saccade", width, height)
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
	sv.SetStruct(&ss.Obj)

	tv := gi.AddNewTabView(split, "tv")

	sc := tv.AddNewTab(gi3d.KiT_Scene, "Scene").(*gi3d.Scene)
	ss.Obj.ConfigScene(sc)

	ss.Obj.ViewImage = tv.AddNewTab(gi.KiT_Bitmap, "Image").(*gi.Bitmap)
	ss.Obj.ViewImage.SetStretchMax()

	ss.TableView = tv.AddNewTab(etview.KiT_TableView, "Table").(*etview.TableView)
	ss.TableView.SetTable(ss.Obj.Sac.Table, nil)

	split.SetSplits(.3, .7)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Obj.Init()
		ss.TableView.UpdateTable()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step env.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Obj.Step()
		ss.TableView.UpdateTable()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step N", Icon: "forward", Tooltip: "Step env N steps.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		for i := 0; i < ss.StepN; i++ {
			ss.Obj.Step()
			vp.FullRender2DTree()
			ss.TableView.UpdateTable()
		}
		vp.SetNeedsFullRender()
	})

	tbar.AddSeparator("run-sep")

	tbar.AddAction(gi.ActOpts{Label: "Run", Icon: "play", Tooltip: "run full set of images and save to file.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		go ss.Obj.Run()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop", Tooltip: "stop running generation.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Obj.Stop()
		vp.SetNeedsFullRender()
	})

	tbar.AddSeparator("env-sep")

	tbar.AddAction(gi.ActOpts{Label: "Env Init", Icon: "reset", Tooltip: "Init Env env.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Obj.Env.Init(0)
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Env Step", Icon: "step-fwd", Tooltip: "Step env.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.Obj.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Obj.Env.Step()
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
