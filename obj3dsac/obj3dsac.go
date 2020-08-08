// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etable"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/gpu"
	"github.com/goki/mat32"
)

// Obj3DSac generates renderings of 3D objects with saccadic eye movements.
// Object is moved around using Sac positions relative to 0,0,0 origin
// and Camera is at 0,0,Z with rotations based on saccade movements.
// Can save a table of saccade data plus image file names to use for
// offline training of models on a cluster using this code, or
// incorporate into an env for more dynamic uses.
type Obj3DSac struct {
	Objs       Obj3D           `desc:"list of 3D objects"`
	Sac        Saccade         `desc:"saccade control"`
	Train      bool            `desc:"if true, use training set of objects, else test"`
	Sequential bool            `desc:"if true, present in sequential order -- else permuted"`
	FOV        float32         `desc:"field of view for camera"`
	ImgSize    image.Point     `desc:"size of image to render"`
	ViewScale  int             `desc:"scale factor for viewing the image"`
	CamPos     mat32.Vec3      `desc:"camera position -- object is positioned around 0,0,0"`
	ZOffScale  float32         `desc:"multiplies the object XYtrg - XYsz to set Zoff to keep general size of objects about the same"`
	ZOffXYtrg  float32         `desc:"target XYsz for z offset scaling"`
	XYPosScale float32         `desc:"multiplier for X,Y positions from saccade"`
	ObjIdx     int             `desc:"index in objs list of current object"`
	Order      []int           `desc:"order to present items (permuted or sequential)"`
	CurObj     string          `inactive:"+" desc:"current object to show"`
	CurCat     string          `inactive:"+" desc:"current category to show"`
	CurXYsz    float32         `inactive:"+" desc:"current object XY size"`
	CurZoff    float32         `inactive:"+" desc:"current Z offset based on XYsz and ZOff*"`
	Image      *gi.Bitmap      `view:"-" desc:"snapshot bitmap view"`
	Scene      *gi3d.Scene     `view:"-" desc:"3D scene"`
	Group      *gi3d.Group     `view:"-" desc:"group holding loaded object"`
	Frame      gpu.Framebuffer `view:"-" desc:"offscreen render buffer"`
}

func (ob *Obj3DSac) Defaults() {
	hdir, _ := os.UserHomeDir()
	path := filepath.Join(hdir, "ccn_images/CU3D_100_plus_models_obj") // downloadable from TODO
	ob.Objs.Path = path
	ob.Objs.NTestPerCat = 2
	ob.Objs.OpenCatProps("cu3d_obj_cat_props.csv")
	ob.Sac.Defaults()
	ob.Sac.TrajLenRange.Set(8, 8)
	ob.Train = true
	ob.FOV = 50
	ob.ImgSize = image.Point{128, 128}
	ob.ViewScale = 4
	ob.CamPos.Z = 3 // set to have object take about 1/2 of width of display overall
	ob.ZOffScale = 2
	ob.ZOffXYtrg = 0.6
	ob.XYPosScale = 1
}

func (ob *Obj3DSac) Config() {
	ob.Sac.Init()
	ob.Objs.Open()
	ob.Init()
}

// ConfigScene must be called with pointer to Scene that is created
// in some form in GUI -- Scene must have access to a Window
func (ob *Obj3DSac) ConfigScene(sc *gi3d.Scene) {
	ob.Scene = sc
	sc.SetStretchMax()
	sc.Defaults()
	sc.BgColor.SetUInt8(103, 176, 255, 255) // sky blue
	sc.Camera.FOV = ob.FOV
	sc.Camera.Pose.Pos = ob.CamPos
	sc.Camera.LookAt(mat32.Vec3Zero, mat32.Vec3Y) // defaults to looking at origin
	dir := gi3d.AddNewDirLight(sc, "dir", 1, gi3d.DirectSun)
	dir.Pos.Set(0, 4, 1) // default: 0,1,1 = above and behind us (we are at 0,0,X)
	ob.Group = gi3d.AddNewGroup(sc, sc, "obj-gp")
}

// ObjList returns the object list to use (Train or Test)
func (ob *Obj3DSac) ObjList() []string {
	if ob.Train {
		return ob.Objs.FlatTrain
	} else {
		return ob.Objs.FlatTest
	}
}

// Init restarts counters
func (ob *Obj3DSac) Init() {
	ob.ObjIdx = -1
	nobj := len(ob.ObjList())
	ob.Order = rand.Perm(nobj)
	ob.Sac.Init()
}

// OpenObj opens object from file path -- relative to Objs.Path
func (ob *Obj3DSac) OpenObj(obj string) error {
	fn := filepath.Join(ob.Objs.Path, obj)
	sc := ob.Scene
	updt := sc.UpdateStart()
	ob.Group.DeleteChildren(true)
	sc.DeleteMeshes()
	sc.DeleteTextures()
	_, err := sc.OpenNewObj(fn, ob.Group)
	if err != nil {
		log.Println(err)
	}
	sc.UpdateEnd(updt)
	sc.UpdateMeshBBox()
	ob.CurXYsz = 0.5 * (ob.Group.MeshBBox.BBox.Max.X + ob.Group.MeshBBox.BBox.Max.Y)
	ob.CurZoff = ob.ZOffScale * (ob.ZOffXYtrg - ob.CurXYsz)
	return err
}

// Render generates image from current object, saving to Image
func (ob *Obj3DSac) Render() error {
	frame := &ob.Frame
	sc := ob.Scene
	if !sc.ActivateOffFrame(frame, "objrend", ob.ImgSize, 4) { // 4 samples
		err := fmt.Errorf("could not activate offscreen framebuffer")
		log.Println(err)
		return err
	}
	if !sc.RenderOffFrame() {
		err := fmt.Errorf("could not render to offscreen framebuffer")
		log.Println(err)
		return err
	}
	(*frame).Rendered()
	// ob.Scene.Render2D()
	// ob.Scene.DirectWinUpload()

	var img image.Image
	oswin.TheApp.RunOnMain(func() {
		tex := (*frame).Texture()
		tex.SetBotZero(true)
		img = tex.GrabImage()
	})
	vwsz := ob.ImgSize.Mul(ob.ViewScale)
	ob.Image.SetImage(img, float32(vwsz.X), float32(vwsz.Y))
	return nil
}

// Position puts object into position according to saccade table
func (ob *Obj3DSac) Position() {
	op := mat32.Vec3{}
	crows := ob.Objs.ObjCatProps.RowsByString("category", ob.CurCat, etable.Equals, etable.UseCase)
	crow := crows[0]
	_ = crow
	// zoff := float32(ob.Objs.ObjCatProps.CellFloat("z_offset", crow))
	// ymirv := ob.Objs.ObjCatProps.CellFloat("y_rot_mirror", crow)
	// ymir := ymirv != 0
	op.Z += ob.CurZoff
	op.X += ob.Sac.ObjPos.X * ob.XYPosScale
	op.Y += ob.Sac.ObjPos.Y * ob.XYPosScale
	ob.Group.Pose.Pos = op
}

// SetObj sets the current obj info based on flat list of objects
func (ob *Obj3DSac) SetObj(list []string) {
	if ob.ObjIdx >= len(list) {
		ob.ObjIdx = 0
		erand.PermuteInts(ob.Order)
	}
	idx := ob.ObjIdx
	if !ob.Sequential {
		idx = ob.Order[ob.ObjIdx]
	}
	ob.CurObj = list[idx]
	ob.CurCat = strings.Split(ob.CurObj, "/")[0]
	ob.OpenObj(ob.CurObj)
}

// Step iterates to next item
func (ob *Obj3DSac) Step() {
	ob.Sac.Step()
	if ob.Sac.NewTraj || ob.ObjIdx < 0 {
		ob.ObjIdx++
		ob.SetObj(ob.ObjList())
	}
	ob.Position()
	ob.Render()
}
