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
	"github.com/goki/gi/gi"
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
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {
	World     FWorld            `desc:"the flat world"`
	StepN     int               `desc:"number of steps to take for StepN button"`
	Table     *etable.Table     `desc:"table recording env"`
	TableView *etview.TableView `view:"-" desc:"the main view"`
	Win       *gi.Window        `view:"-" desc:"main GUI window"`
	ToolBar   *gi.ToolBar       `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.StepN = 8
	ss.World.Config(1000)
	ss.World.Init(0)

	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"DepthView", etensor.FLOAT32, ss.World.CurStates["DepthView"].Shape.Shp, nil},
		{"Fovea", etensor.FLOAT32, ss.World.CurStates["Fovea"].Shape.Shp, nil},
		{"ProxSoma", etensor.FLOAT32, ss.World.CurStates["ProxSoma"].Shape.Shp, nil},
		{"Vestibular", etensor.FLOAT32, ss.World.CurStates["Vestibular"].Shape.Shp, nil},
		{"Inters", etensor.FLOAT32, ss.World.CurStates["Inters"].Shape.Shp, nil},
	}
	ss.Table = etable.NewTable("input")
	ss.Table.SetFromSchema(sch, 1)
	ss.Table.SetMetaData("TrialName:width", "40")
}

// Step takes one step and records in table
func (ss *Sim) Step() {
	ss.World.Step()
	for i := 1; i < ss.Table.NumCols(); i++ {
		cnm := ss.Table.ColNames[i]
		inp := ss.World.State(cnm)
		ss.Table.SetCellTensor(cnm, 0, inp)
	}
	ss.Table.SetCellString("TrialName", 0, ss.World.String())
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

	// sc := tv.AddNewTab(gi3d.KiT_Scene, "Scene").(*gi3d.Scene)
	// ss.World.ConfigScene(sc)

	// ss.World.ViewImage = tv.AddNewTab(gi.KiT_Bitmap, "Image").(*gi.Bitmap)
	// ss.World.ViewImage.SetStretchMax()

	ss.TableView = tv.AddNewTab(etview.KiT_TableView, "Table").(*etview.TableView)
	ss.TableView.SetTable(ss.Table, nil)

	split.SetSplits(.3, .7)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "reset", Tooltip: "Init env."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.World.Init(0)
			ss.TableView.UpdateTable()
			vp.SetNeedsFullRender()
		})

	tbar.AddAction(gi.ActOpts{Label: "Left", Icon: "wedge-left", Tooltip: "Rotate Left"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Left()
			ss.TableView.UpdateTable()
			vp.SetNeedsFullRender()
		})

	tbar.AddAction(gi.ActOpts{Label: "Right", Icon: "wedge-right", Tooltip: "Rotate Right"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Right()
			ss.TableView.UpdateTable()
			vp.SetNeedsFullRender()
		})

	tbar.AddAction(gi.ActOpts{Label: "Forward", Icon: "wedge-up", Tooltip: "Step Forward"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Forward()
			ss.TableView.UpdateTable()
			vp.SetNeedsFullRender()
		})

	tbar.AddAction(gi.ActOpts{Label: "Backward", Icon: "wedge-down", Tooltip: "Step Backward"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Backward()
			ss.TableView.UpdateTable()
			vp.SetNeedsFullRender()
		})

	tbar.AddSeparator("sep-file")

	tbar.AddAction(gi.ActOpts{Label: "Open World", Icon: "file-open", Tooltip: "Open World from .tsv file"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(&ss.World, "OpenWorld", vp)
		})

	tbar.AddAction(gi.ActOpts{Label: "Save World", Icon: "file-save", Tooltip: "Save World to .tsv file"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(&ss.World, "SaveWorld", vp)
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
