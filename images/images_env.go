// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/anthonynsimon/bild/transform"
	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
)

// ImagesEnv provides the rendered results of the Obj3D + Saccade generator.
type ImagesEnv struct {
	Nm     string  `desc:"name of this environment"`
	Dsc    string  `desc:"description of this environment"`
	Test   bool    `desc:"present test items, else train"`
	Images Images  `desc:"images list"`
	V1m16  Vis     `desc:"v1 16deg medium resolution filtering of image -- V1AllTsr has result"`
	V1h16  Vis     `desc:"v1 16deg higher resolution filtering of image -- V1AllTsr has result"`
	V1m8   Vis     `desc:"v1 8deg medium resolution filtering of image -- V1AllTsr has result"`
	V1h8   Vis     `desc:"v1 8deg higher resolution filtering of image -- V1AllTsr has result"`
	Order  []int   `desc:"order of images to present"`
	Run    env.Ctr `view:"inline" desc:"current run of model as provided during Init"`
	Epoch  env.Ctr `view:"inline" desc:"arbitrary aggregation of trials, for stats etc"`
	Trial  env.Ctr `view:"inline" desc:"each object trajectory is one trial"`
	Row    env.Ctr `view:"inline" desc:"row of item list  -- this is actual counter driving everything"`
	CurCat string  `desc:"current category"`
	CurImg string  `desc:"current image"`

	Image image.Image `view:"-" desc:"rendered image as loaded"`
}

func (ev *ImagesEnv) Name() string { return ev.Nm }
func (ev *ImagesEnv) Desc() string { return ev.Dsc }

func (ev *ImagesEnv) Validate() error {
	return nil
}

func (ev *ImagesEnv) Defaults() {
	ev.V1m16.Defaults(24, 8)
	ev.V1h16.Defaults(12, 4)
	ev.V1m8.Defaults(12, 4)
	ev.V1m8.V1sGeom.Border = image.Point{38, 38}
	ev.V1h8.Defaults(6, 2)
	ev.V1h8.V1sGeom.Border = image.Point{38, 38}
}

// ImageList returns the list of images -- train or test
func (ev *ImagesEnv) ImageList() []string {
	if ev.Test {
		return ev.Images.FlatTest
	}
	return ev.Images.FlatTrain
}

func (ev *ImagesEnv) Init(run int) {
	ev.Run.Scale = env.Run
	ev.Epoch.Scale = env.Epoch
	ev.Trial.Scale = env.Trial
	ev.Row.Scale = env.Tick
	ev.Run.Init()
	ev.Epoch.Init()
	ev.Trial.Init()
	ev.Run.Cur = run
	ev.Row.Cur = -1 // init state -- key so that first Step() = 0
	ev.Row.Max = len(ev.ImageList())
	ev.Order = rand.Perm(ev.Row.Max)
}

// SaveListJSON saves flat string list to a JSON-formatted file.
func SaveListJSON(list []string, filename string) error {
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		log.Println(err)
	}
	return err
}

// OpenListJSON opens flat string list from a JSON-formatted file.
func OpenListJSON(list *[]string, filename string) error {
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, list)
}

// SaveList2JSON saves double-string list to a JSON-formatted file.
func SaveList2JSON(list [][]string, filename string) error {
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		log.Println(err)
	}
	return err
}

// OpenList2JSON opens double-string list from a JSON-formatted file.
func OpenList2JSON(list *[][]string, filename string) error {
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, list)
}

// OpenConfig opens saved configuration for current images
func (ev *ImagesEnv) OpenConfig() bool {
	cfnm := fmt.Sprintf("%s_cats.json", ev.Nm)
	tsfnm := fmt.Sprintf("%s_ntest%d_tst.json", ev.Nm, ev.Images.NTestPerCat)
	trfnm := fmt.Sprintf("%s_ntest%d_trn.json", ev.Nm, ev.Images.NTestPerCat)
	_, err := os.Stat(tsfnm)
	if !os.IsNotExist(err) {
		OpenListJSON(&ev.Images.Cats, cfnm)
		OpenList2JSON(&ev.Images.ImagesTest, tsfnm)
		OpenList2JSON(&ev.Images.ImagesTrain, trfnm)
		ev.Images.Flats()
		return true
	}
	return false
}

// SaveConfig saves configuration for current images
func (ev *ImagesEnv) SaveConfig() {
	cfnm := fmt.Sprintf("%s_cats.json", ev.Nm)
	tsfnm := fmt.Sprintf("%s_ntest%d_tst.json", ev.Nm, ev.Images.NTestPerCat)
	trfnm := fmt.Sprintf("%s_ntest%d_trn.json", ev.Nm, ev.Images.NTestPerCat)
	SaveListJSON(ev.Images.Cats, cfnm)
	SaveList2JSON(ev.Images.ImagesTest, tsfnm)
	SaveList2JSON(ev.Images.ImagesTrain, trfnm)
}

// CurImage returns current image based on row and
func (ev *ImagesEnv) CurImage() string {
	il := ev.ImageList()
	sz := len(il)
	if len(ev.Order) != sz {
		ev.Order = rand.Perm(ev.Row.Max)
	}
	if ev.Row.Cur >= sz {
		ev.Row.Max = sz
		ev.Row.Cur = 0
		erand.PermuteInts(ev.Order)
	}
	r := ev.Row.Cur
	if r < 0 {
		r = 0
	}
	i := ev.Order[r]
	ev.CurImg = il[i]
	ev.CurCat = ev.Images.Cat(ev.CurImg)
	return ev.CurImg
}

// OpenImage opens current image
func (ev *ImagesEnv) OpenImage() error {
	img := ev.CurImage()
	fnm := filepath.Join(ev.Images.Path, img)
	var err error
	ev.Image, err = gi.OpenImage(fnm)
	if err != nil {
		log.Println(err)
	}
	return err
}

// FilterImage opens and filters current image
func (ev *ImagesEnv) FilterImage() error {
	err := ev.OpenImage()
	if err != nil {
		fmt.Println(err)
		return err
	}
	// todo: transform image
	// resize once for both..
	tsz := ev.V1m16.ImgSize
	isz := ev.Image.Bounds().Size()
	if isz != tsz {
		ev.Image = transform.Resize(ev.Image, tsz.X, tsz.Y, transform.Linear)
	}
	ev.V1m16.Filter(ev.Image)
	ev.V1h16.Filter(ev.Image)
	ev.V1m8.Filter(ev.Image)
	ev.V1h8.Filter(ev.Image)
	return nil
}

func (ev *ImagesEnv) String() string {
	return fmt.Sprintf("%s:%s_%d", ev.CurCat, ev.CurImage, ev.Trial.Cur)
}

func (ev *ImagesEnv) Step() bool {
	ev.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	ev.Row.Incr()   // auto-rotates
	if ev.Trial.Incr() {
		ev.Epoch.Incr()
	}
	ev.FilterImage()
	return true
}

func (ev *ImagesEnv) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return ev.Run.Query()
	case env.Epoch:
		return ev.Epoch.Query()
	case env.Trial:
		return ev.Trial.Query()
	}
	return -1, -1, false
}

func (ev *ImagesEnv) State(element string) etensor.Tensor {
	switch element {
	case "V1m16":
		return &ev.V1m16.V1AllTsr
	case "V1h16":
		return &ev.V1h16.V1AllTsr
	case "V1m8":
		return &ev.V1m8.V1AllTsr
	case "V1h8":
		return &ev.V1h8.V1AllTsr
	}
	return nil
}

func (ev *ImagesEnv) Action(element string, input etensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ env.Env = (*ImagesEnv)(nil)
