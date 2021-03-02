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
	"strconv"
	"strings"
	"flag"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/gpu"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
)

// Obj3DSac generates renderings of 3D objects with saccadic eye movements.
// Object is moved around using Sac positions relative to 0,0,0 origin
// and Camera is at 0,0,Z with rotations based on saccade movements.
// Can save a table of saccade data plus image file names to use for
// offline training of models on a cluster using this code, or
// incorporate into an env for more dynamic uses.
type Obj3DSac struct {
	Objs       Obj3D         `desc:"list of 3D objects"`
	Sac        Saccade       `desc:"saccade control"`
	// TODO first pass
	// Env        Obj3DSacEnv   `desc:"environment that loads rendered images"`

	RandomAction bool        `desc:"whether to interally generate random saccades or to use externally generated saccades as action inputs"`
	Agent      BaseAgent    `view:"-" desc:"the agent generating actions"`
	SacActionPop popcode.TwoD `desc:"popcode for decoding saccade actions"`
	SaveFiles  bool          `desc:"if true, save images (in epoch-wise subdirs) and data.tsv file with saccade position data and image name, to images dir"`
	NTrials    int           `desc:"number of trials per epoch, for saving"`
	NEpcs      int           `desc:"number of epochs"`
	Table      *etable.Table `desc:"if saving, this is the trial-by-trial data"`
	Train      bool          `desc:"if true, use training set of objects, else test"`
	Sequential bool          `desc:"if true, present in sequential order -- else permuted"`
	FOV        float32       `desc:"field of view for camera"`
	ImgSize    image.Point   `desc:"size of image to render"`
	ViewScale  int           `desc:"scale factor for viewing the image"`
	CamPos     mat32.Vec3    `desc:"camera position -- object is positioned around 0,0,0"`
	ZOffScale  float32       `desc:"multiplies the object XYtrg - XYsz to set Zoff to keep general size of objects about the same"`
	ZOffXYtrg  float32       `desc:"target XYsz for z offset scaling"`
	XYPosScale float32       `desc:"multiplier for X,Y positions from saccade"`
	// TODO first pass
	Rot3D      []mat32.Vec3  `desc:"how much to rotate along each axis, in degrees per step"`
	ObjIdx     int           `desc:"index in objs list Order of first current object in order"`
	Order      []int         `desc:"order to present items (permuted or sequential)"`

	// TODO first pass
	NObjScene int        `inactive:"+" desc:"number of objects to show simultaneously"`
	CurObj  []string     `inactive:"+" desc:"current objects to show"`
	CurCat  []string     `inactive:"+" desc:"current categories to show"`
	CurXYsz []float32    `inactive:"+" desc:"current object XY sizes"`
	CurZoff []float32    `inactive:"+" desc:"current Z offsets based on XYsz and ZOff*"`
	InitRot []mat32.Vec3 `inactive:"+" desc:"initial euler 3D rotations, in degrees"`
	RotVel  []mat32.Vec3 `inactive:"+" desc:"3D rotational velocities (degrees per step) for current objects"`
	CurRot  []mat32.Vec3 `inactive:"+" desc:"current euler rotations"`

	Trial   env.Ctr    `inactive:"+" desc:"current trial, for saving"`
	Epoch   env.Ctr    `inactive:"+" desc:"current epoch, for saving"`
	SaveDir string     `inactive:"+" desc:"name of current directory to save files into (images/train or images/test)"`
	ImgFile string     `inactive:"+" desc:"name of image file"`

	Image     image.Image     `view:"-" desc:"current rendered image in specified size"`
	ViewImage *gi.Bitmap      `view:"-" desc:"View of image (scaled up) as a bitmap"`
	Scene     *gi3d.Scene     `view:"-" desc:"3D scene"`

	// TODO first pass
	Group     []*gi3d.Group   `view:"-" desc:"group holding loaded objects"`
	Frame     gpu.Framebuffer `view:"-" desc:"offscreen render buffer"`
	File      *os.File        `view:"-" desc:"save file"`
	StopNow   bool            `view:"-" desc:"flag to stop running"`
	IsRunning bool            `view:"-" desc:"true when running"`
}

func (ob *Obj3DSac) Defaults() {
	// hdir, _ := os.UserHomeDir()
	// path := filepath.Join(hdir, "ccn_images/CU3D_100_plus_models_obj") // downloadable from TODO
	// ob.Objs.Path = path
	ob.Objs.NTestPerCat = 2
	// ob.Objs.OpenCatProps("cu3d_obj_cat_props.csv")
	ob.CmdArgs()
	ob.Sac.Defaults()
	ob.Sac.NObjScene = ob.NObjScene
	ob.Sac.NObjSacLim = 1
	ob.Sac.TrajLenRange.Set(8, 8)
	ob.NTrials = 64
	ob.NEpcs = 1000
	ob.Train = true
	ob.FOV = 50
	ob.ImgSize = image.Point{256, 256}
	ob.ViewScale = 2
	ob.CamPos.Z = 3 // set to have object take about 1/2 of width of display overall
	ob.ZOffScale = 2
	ob.ZOffXYtrg = 0.6
	ob.XYPosScale = 1.5
	// TODO first pass
	ob.Rot3D = make([]mat32.Vec3, ob.NObjScene)
	ob.CurObj = make([]string, ob.NObjScene)
	ob.CurCat = make([]string, ob.NObjScene)
	ob.CurXYsz = make([]float32, ob.NObjScene)
	ob.CurZoff = make([]float32, ob.NObjScene)
	ob.InitRot = make([]mat32.Vec3, ob.NObjScene)
	ob.RotVel = make([]mat32.Vec3, ob.NObjScene)
	ob.CurRot = make([]mat32.Vec3, ob.NObjScene)
	ob.Group = make([]*gi3d.Group, ob.NObjScene)

	for i := 0; i < ob.NObjScene; i++ {
		ob.Rot3D[i].Set(0, 5, 0.5)
	}

	if !ob.RandomAction {
		ob.SacActionPop.Defaults()
		ob.SacActionPop.Min.Set(-0.45, -0.45)
		ob.SacActionPop.Max.Set(0.45, 0.45)
		ob.Sac.RandomAction = false
	}

	ob.Trial.Scale = env.Trial
	ob.Epoch.Scale = env.Epoch
}

func (ob *Obj3DSac) CmdArgs() {
	hdir, _ := os.UserHomeDir()
	path_default := filepath.Join(hdir, "ccn_images/CU3D_100_plus_models_obj")

	// flags can't be defined more than once (e.g. such as separately in TrainEnv and TestEnv)
	if flag.Lookup("obj_path") == nil {
		flag.StringVar(&ob.Objs.Path, "obj_path", path_default, "Path to directory containing .obj files in subdirectories of categories")
	} else {
		ob.Objs.Path = flag.Lookup("obj_path").Value.(flag.Getter).Get().(string)
	}

	var cat_path string
	if flag.Lookup("obj_cat_csv_path") == nil {
		flag.StringVar(&cat_path, "obj_cat_csv_path", "cu3d_obj_cat_props.csv", "Path to CSV containing object categories with Z-offset and y-rotated/mirrored values")
	} else {
		cat_path = flag.Lookup("obj_cat_csv_path").Value.(flag.Getter).Get().(string)
	}

	if flag.Lookup("n_obj_scene") == nil {
		flag.IntVar(&ob.NObjScene, "n_obj_scene", 1, "Number of objects to simultaneously include in the scene")
	} else {
		ob.NObjScene = flag.Lookup("n_obj_scene").Value.(flag.Getter).Get().(int)
	}

	if flag.Lookup("rand_act") == nil {
		flag.BoolVar(&ob.RandomAction, "rand_act", false, "Whether to use random actions or an external policy to generate actions.")
	} else {
		ob.RandomAction = flag.Lookup("rand_act").Value.(flag.Getter).Get().(bool)
	}
	flag.Parse()
	ob.Objs.OpenCatProps(cat_path)
}

func (ob *Obj3DSac) Config() {
	ob.Sac.Init()
	ob.Objs.Open()
	// must uncomment for standard testing with CU100-3D data set
	ob.Objs.DeleteCats(ObjsBigSlow) // avoid!
	ob.Objs.SelectCats(Objs20)
	ob.Objs.SelectObjs(Objs20orig)
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
	dir.Pos.Set(0, 1, 1) // default: 0,1,1 = above and behind us (we are at 0,0,X)
	// dir = gi3d.AddNewDirLight(sc, "dir2", 1, gi3d.DirectSun)
	// dir.Pos.Set(0, 1, 0) // directly above
	for i := 0; i < ob.NObjScene; i++ {
		ob.Group[i] = gi3d.AddNewGroup(sc, sc, "obj-gp")
	}
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

// TODO first pass

// OpenObj opens object from file path -- relative to Objs.Path
// func (ob *Obj3DSac) OpenObj(obj string) error {
func (ob *Obj3DSac) OpenObj(objs []string) error {
	// fn := filepath.Join(ob.Objs.Path, obj)
	fns := make([]string, len(objs))
	for i, obj := range objs {
		fns[i] = filepath.Join(ob.Objs.Path, obj)
	}

	sc := ob.Scene
	updt := sc.UpdateStart()
	var err error
	for i := 0; i < ob.NObjScene; i++ {
		ob.Group[i].DeleteChildren(true)
	}
	sc.DeleteMeshes()
	sc.DeleteTextures()
	ki.DelMgr.DestroyDeleted() // this is actually essential to prevent leaking memory!
	// fmt.Printf("Epc: %d \t Trial: %d \t Opening object: %s\n", ob.Epoch.Cur, ob.Trial.Cur, fn)
	fmt.Printf("Epc: %d \t Trial: %d \t Opening object: %s\n", ob.Epoch.Cur, ob.Trial.Cur, fns)

	for i, fn := range fns {
		_, err = sc.OpenNewObj(fn, ob.Group[i])
		if err != nil {
			log.Println(err)
		}

		sc.UpdateEnd(updt)
		sc.UpdateMeshBBox()
		ob.CurXYsz[i] = 0.5 * (ob.Group[i].MeshBBox.BBox.Max.X + ob.Group[i].MeshBBox.BBox.Max.Y)
		ob.CurZoff[i] = ob.ZOffScale * (ob.ZOffXYtrg - ob.CurXYsz[i])
		crows := ob.Objs.ObjCatProps.RowsByString("category", ob.CurCat[i], etable.Equals, etable.UseCase)
		crow := crows[0]
		// zoff := float32(ob.Objs.ObjCatProps.CellFloat("z_offset", crow))
		ymirv := ob.Objs.ObjCatProps.CellFloat("y_rot_mirror", crow)
		ymir := ymirv != 0
		yflip := erand.BoolProb(.5, -1)
		if ymir && yflip {
			ob.InitRot[i].Y = 180
		} else {
			ob.InitRot[i].Y = 0
		}
		ob.RotVel[i].Z = -ob.Rot3D[i].Z + 2*ob.Rot3D[i].Z*rand.Float32()
		ob.RotVel[i].Y = -ob.Rot3D[i].Y + 2*ob.Rot3D[i].Y*rand.Float32()
		if err != nil {
			return err
		}
	}
	return err
}

// Render generates image from current object, saving to Image
func (ob *Obj3DSac) Render() error {
	frame := &ob.Frame
	sc := ob.Scene
	err := sc.ActivateOffFrame(frame, "objrend", ob.ImgSize, 4)
	if err != nil {
		return err
	}
	sc.RenderOffFrame()
	(*frame).Rendered()

	oswin.TheApp.RunOnMain(func() {
		tex := (*frame).Texture()
		tex.SetBotZero(true)
		ob.Image = tex.GrabImage()
	})
	if ob.ViewImage != nil {
		vwsz := ob.ImgSize.Mul(ob.ViewScale)
		ob.ViewImage.SetImage(ob.Image, float32(vwsz.X), float32(vwsz.Y))
	}
	return nil
}

// Position puts objects into position according to saccade table
// TODO first pass
func (ob *Obj3DSac) Position() {
	for i := 0; i < ob.NObjScene; i++ {
		op := mat32.Vec3{}
		op.Z += ob.CurZoff[i]
		op.X += ob.Sac.ObjPos[i].X * ob.XYPosScale
		op.Y += ob.Sac.ObjPos[i].Y * ob.XYPosScale
		ob.Group[i].Pose.Pos = op
		ob.CurRot[i] = ob.InitRot[i].Add(ob.RotVel[i].MulScalar(float32(ob.Sac.Tick.Cur)))
		ob.Group[i].Pose.SetEulerRotation(ob.CurRot[i].X, ob.CurRot[i].Y, ob.CurRot[i].Z)
	}
}

// Fixate moves rendering camera to fixate on eye position
func (ob *Obj3DSac) Fixate() {
	sc := ob.Scene
	trg := mat32.Vec3{}
	trg.X = ob.Sac.EyePos.X * ob.XYPosScale
	trg.Y = ob.Sac.EyePos.Y * ob.XYPosScale
	sc.Camera.LookAt(trg, mat32.Vec3Y)
}

// SetObj sets the current obj info based on flat list of objects
func (ob *Obj3DSac) SetObj(list []string) {
	// TODO first pass
	if ob.ObjIdx + ob.NObjScene - 1 >= len(list) {
		ob.ObjIdx = 0
		erand.PermuteInts(ob.Order)
	}

	for i := 0; i < ob.NObjScene; i++ {
		idx := ob.ObjIdx + i
		if !ob.Sequential {
			idx = ob.Order[idx]
		}
		ob.CurObj[i] = list[idx]
		ob.CurCat[i] = strings.Split(ob.CurObj[i], "/")[0]
	}
	ob.OpenObj(ob.CurObj)
}

// Step iterates to next item
func (ob *Obj3DSac) Step() {
	ob.Sac.Step()
	// TODO 
	if ob.Sac.NewTraj || ob.ObjIdx < 0 {
		ob.ObjIdx += ob.NObjScene
		ob.SetObj(ob.ObjList()) // wraps objidx
		if ob.Trial.Incr() {
			if ob.Epoch.Incr() {
				ob.Stop()
				return
			}
		}
	}
	ob.Position()
	ob.Fixate()
	err := ob.Render()
	if err != nil {
		ob.Stop()
	}
	ob.SaveTick()
}

// TODO ask about switching environment API over to this
func (ob *Obj3DSac) Action(action map[string]etensor.Tensor) {
	a := mat32.Vec2{}
	a, _ = ob.SacActionPop.Decode(action["SacPlan"])
	ob.Sac.CondSetSacPlan(a)
}

// Run runs full set of Save trials / epochs
func (ob *Obj3DSac) Run() {
	ob.Trial.Max = ob.NTrials
	ob.Epoch.Max = ob.NEpcs
	ob.Trial.Init()
	ob.Trial.Cur = -1 // gets inc to 0 at start
	ob.Epoch.Init()

	ob.Table = &etable.Table{}
	ob.ConfigTable(ob.Table)
	ob.Table.SetNumRows(1) // just re-use same row.. fine..

	if ob.Train {
		ob.SaveDir = "images/train"
	} else {
		ob.SaveDir = "images/test"
	}

	var err error
	os.MkdirAll(ob.SaveDir, 0755)

	var action map[string]etensor.Tensor
	var state map[string]etensor.Tensor

	ob.File, err = os.Create(filepath.Join(ob.SaveDir, "data.tsv"))
	if err != nil {
		log.Println(err)
		return
	}

	if ob.Train {
		SaveListJSON(ob.Objs.FlatTrain, filepath.Join(ob.SaveDir, "objs.json"))
	} else {
		SaveListJSON(ob.Objs.FlatTest, filepath.Join(ob.SaveDir, "objs.json"))
	}
	SaveListJSON(ob.Objs.Cats, filepath.Join(ob.SaveDir, "cats.json"))

	ob.SaveFiles = true
	ob.IsRunning = true

	vp := ob.Scene.Win.WinViewport2D()

	for {
		if !ob.RandomAction {
			action = ob.Agent.Step(state)  // TODO implement ob.State() to give state tensor as desired for an agent
			ob.Action(action)
			ob.Step()
		} else {
			ob.Step()
		}
		vp.FullRender2DTree()  // useful for debugging
		if ob.StopNow {
			ob.StopNow = false
			break
		}
	}
	ob.IsRunning = false
	vp.FullRender2DTree()
}

// Stop tells the sim to stop running
func (ob *Obj3DSac) Stop() {
	ob.StopNow = true
}

func (ob *Obj3DSac) ConfigTable(dt *etable.Table) {
	dt.SetMetaData("name", "Obj3DSacTable")
	dt.SetMetaData("desc", "table of obj3d data")
	dt.SetMetaData("read-only", "true")
		dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	// TODO first pass, not sure if etensor.STRING types can have cell shapes that aren't nil
	sch := etable.Schema{
		{"Epoch", etensor.INT64, nil, nil},
		{"Trial", etensor.INT64, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"SacTick", etensor.INT64, nil, nil},
		{"Cat", etensor.STRING, []int{ob.NObjScene}, nil},
		{"Obj", etensor.STRING, []int{ob.NObjScene}, nil},
		{"ImgFile", etensor.STRING, nil, nil},
		{"ObjPos", etensor.FLOAT32, []int{ob.NObjScene, 2}, nil},
		{"ObjViewPos", etensor.FLOAT32, []int{ob.NObjScene, 2}, nil},
		{"ObjVel", etensor.FLOAT32, []int{ob.NObjScene, 2}, nil},
		{"ObjPosNext", etensor.FLOAT32, []int{ob.NObjScene, 2}, nil},
		{"ObjRot", etensor.FLOAT32, []int{ob.NObjScene, 3}, nil},
		{"EyePos", etensor.FLOAT32, []int{2}, nil},
		{"SacPlan", etensor.FLOAT32, []int{2}, nil},
		{"Saccade", etensor.FLOAT32, []int{2}, nil},
	}
	dt.SetFromSchema(sch, 0)
}

// SaveTick saves the current tick, if saving
func (ob *Obj3DSac) SaveTick() {
	if !ob.SaveFiles || ob.Table == nil {
		return
	}
	sc := &ob.Sac

	epc := ob.Epoch.Cur
	trl := ob.Trial.Cur
	tick := sc.Tick.Cur

	objs := make([]string, ob.NObjScene)
	for i := range ob.CurObj {
		objs[i] = strings.Split(ob.CurObj[i], "/")[1]
	}

	epcdir := fmt.Sprintf("epc_%04d", epc)
	imgdir := filepath.Join(ob.SaveDir, epcdir)
	os.MkdirAll(imgdir, 0755)

	ob.ImgFile = fmt.Sprintf("%s/img_%04d_%03d_%d.jpg", epcdir, epc, trl, tick)

	gi.SaveImage(filepath.Join(ob.SaveDir, ob.ImgFile), ob.Image)

	dt := ob.Table
	row := 0

	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("Trial", row, float64(trl))
	dt.SetCellFloat("Tick", row, float64(tick))
	dt.SetCellFloat("SacTick", row, float64(sc.SacTick.Cur))

	// TODO first pass
	cat_col := dt.ColByName("Cat") 
	obj_col := dt.ColByName("Obj") 

	dt.SetCellString("ImgFile", row, ob.ImgFile)

	// this is from saccade.go:
	posTsr := etensor.NewFloat64([]int{sc.NObjScene, 2}, nil, nil)
	viewPosTsr := etensor.NewFloat64([]int{sc.NObjScene, 2}, nil, nil)
	velTsr := etensor.NewFloat64([]int{sc.NObjScene, 2}, nil, nil)
	posNextTsr := etensor.NewFloat64([]int{sc.NObjScene, 2}, nil, nil)
	rotTsr := etensor.NewFloat64([]int{sc.NObjScene, 3}, nil, nil)

	for i := 0; i < sc.NObjScene; i++ {
		cat_col.SetString([]int{row, i}, ob.CurCat[i])
		obj_col.SetString([]int{row, i}, objs[i])

		posTsr.SetFloat([]int{i, 0}, float64(sc.ObjPos[i].X))
		posTsr.SetFloat([]int{i, 1}, float64(sc.ObjPos[i].Y))

		viewPosTsr.SetFloat([]int{i, 0}, float64(sc.ObjViewPos[i].X))
		viewPosTsr.SetFloat([]int{i, 1}, float64(sc.ObjViewPos[i].Y))

		velTsr.SetFloat([]int{i, 0}, float64(sc.ObjVel[i].X))
		velTsr.SetFloat([]int{i, 1}, float64(sc.ObjVel[i].Y))

		posNextTsr.SetFloat([]int{i, 0}, float64(sc.ObjPosNext[i].X))
		posNextTsr.SetFloat([]int{i, 1}, float64(sc.ObjPosNext[i].Y))

		rotTsr.SetFloat([]int{i, 0}, float64(ob.CurRot[i].X))
		rotTsr.SetFloat([]int{i, 1}, float64(ob.CurRot[i].Y))
		rotTsr.SetFloat([]int{i, 2}, float64(ob.CurRot[i].Z))
	}

	dt.SetCellTensor("ObjPos", row, posTsr)
	dt.SetCellTensor("ObjViewPos", row, viewPosTsr)
	dt.SetCellTensor("ObjVel", row, velTsr)
	dt.SetCellTensor("ObjPosNext", row, posNextTsr)
	dt.SetCellTensor("ObjRot", row, rotTsr)

	dt.SetCellTensorFloat1D("EyePos", row, 0, float64(sc.EyePos.X))
	dt.SetCellTensorFloat1D("EyePos", row, 1, float64(sc.EyePos.Y))
	dt.SetCellTensorFloat1D("SacPlan", row, 0, float64(sc.SacPlan.X))
	dt.SetCellTensorFloat1D("SacPlan", row, 1, float64(sc.SacPlan.Y))
	dt.SetCellTensorFloat1D("Saccade", row, 0, float64(sc.Saccade.X))
	dt.SetCellTensorFloat1D("Saccade", row, 1, float64(sc.Saccade.Y))

	if ob.File != nil {
		if trl == 0 && epc == 0 && tick == 0 {
			dt.WriteCSVHeaders(ob.File, etable.Tab)
		}
		dt.WriteCSVRow(ob.File, row, etable.Tab)
	}
}
