// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// main for GUI interaction with Env for testing
package main

import (
	"os"
	"path/filepath"

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
	Images  ImagesEnv         `desc:"the env item"`
	View    *etview.TableView `view:"-" desc:"the main view"`
	Win     *gi.Window        `view:"-" desc:"main GUI window"`
	ToolBar *gi.ToolBar       `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	hdir, _ := os.UserHomeDir()
	// see README.md for download info
	path := filepath.Join(hdir, "ccn_images/CU3D_100_plus_renders")
	ss.Images.Defaults()
	ss.Images.Nm = "cu3d100plus"
	ss.Images.Images.NTestPerCat = 2
	ss.Images.Images.SplitByItm = true
	// switch the next two lines to use saved splits
	// ss.Images.Images.SetPath(path, []string{".png"}, "_")
	// ss.Images.OpenConfig()
	ss.Images.Images.OpenPath(path, []string{".png"}, "_")
	ss.Images.SaveConfig()
	ss.Images.Init(0)
	ss.Images.Step()
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
	sv.SetStruct(&ss.Images)

	tv := gi.AddNewTabView(split, "tv")

	bimg := tv.AddNewTab(gi.KiT_Bitmap, "Image").(*gi.Bitmap)
	bimg.SetStretchMax()
	bimg.SetImage(ss.Images.Image, 1024, 1024)

	v1h16 := tv.AddNewTab(etview.KiT_TensorGrid, "V1h16").(*etview.TensorGrid)
	v1h16.SetStretchMax()
	v1h16.SetTensor(&ss.Images.V1h16.V1AllTsr)

	v1m16 := tv.AddNewTab(etview.KiT_TensorGrid, "V1m16").(*etview.TensorGrid)
	v1m16.SetStretchMax()
	v1m16.SetTensor(&ss.Images.V1m16.V1AllTsr)

	v1h8 := tv.AddNewTab(etview.KiT_TensorGrid, "V1h8").(*etview.TensorGrid)
	v1h8.SetStretchMax()
	v1h8.SetTensor(&ss.Images.V1h8.V1AllTsr)

	v1m8 := tv.AddNewTab(etview.KiT_TensorGrid, "V1m8").(*etview.TensorGrid)
	v1m8.SetStretchMax()
	v1m8.SetTensor(&ss.Images.V1m8.V1AllTsr)

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "process next input"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Images.Step()
		bimg.SetImage(ss.Images.Image, 1024, 1024)
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "XForm", Icon: "step-fwd", Tooltip: "transform image according to current settings"}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Images.FilterImage()
		bimg.SetImage(ss.Images.Image, 1024, 1024)
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
