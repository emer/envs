// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goki/ki/dirs"
)

// Images implements management of lists of image files,
// with category names or organized in directories by category.
type Images struct {
	Path        string     `desc:"path to image files -- this should point to a directory that has files or subdirectories that then have image files in them"`
	Exts        []string   `desc:"extensions of image files to find (lowercase)"`
	CatSep      string     `desc:"separator in file name for category label -- if empty then must have subdirs"`
	SplitByItm  bool       `desc:"split by item -- each file name has an item label after CatSep"`
	NTestPerCat int        `desc:"number of testing images per category -- if SplitByItem images are split by item id"`
	Cats        []string   `desc:"list of image categories"`
	ImagesAll   [][]string `desc:"full list of images, organized by category (directory) and then filename"`
	ImagesTrain [][]string `desc:"list of training images, organized by category (directory) and then filename"`
	ImagesTest  [][]string `desc:"list of testing images, organized by category (directory) and then filename"`
	FlatAll     []string   `desc:"flat list of all images, as cat/filename.ext -- Flats() makes from above"`
	FlatTrain   []string   `desc:"flat list of all training images, as cat/filename.ext -- Flats() makes from above"`
	FlatTest    []string   `desc:"flat list of all testing images, as cat/filename.ext -- Flats() makes from above"`
}

// OpenPath opens list of images at given path, with given extensions
func (im *Images) OpenPath(path string, exts []string, catsep string) error {
	im.Path = path
	im.Exts = exts
	im.CatSep = catsep
	if im.CatSep == "" {
		return im.OpenDirs()
	}
	return im.OpenNames()
}

// OpenDirs opens images at Path with subdirs for category names
func (im *Images) OpenDirs() error {
	im.Cats = dirs.Dirs(im.Path)
	nc := len(im.Cats)
	if nc == 0 {
		err := fmt.Errorf("Images.OpenDirs() -- no directories for categories in: %s", im.Path)
		log.Println(err)
		return err
	}
	im.ImagesAll = make([][]string, nc)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := im.Cats[ci]
		cp := filepath.Join(im.Path, cat)
		fls := dirs.ExtFileNames(cp, im.Exts)
		if len(fls) == 0 {
			im.Cats = append(im.Cats[:ci], im.Cats[ci+1:]...)
			im.ImagesAll = append(im.ImagesAll[:ci], im.ImagesAll[ci+1:]...)
			continue
		}
		im.ImagesAll[ci] = fls
	}
	im.Split()
	return nil
}

func (im *Images) Cat(f string) string {
	if im.CatSep == "" {
		dir, _ := filepath.Split(f)
		return dir
	}
	i := strings.Index(f, im.CatSep)
	return f[:i]
}

func (im *Images) Item(f string) string {
	i := strings.Index(f, im.CatSep)
	rf := f[i+1:]
	i = strings.Index(rf, im.CatSep)
	return rf[:i]
}

// OpenNames opens images at Path with category names in file names
func (im *Images) OpenNames() error {
	fls := dirs.ExtFileNames(im.Path, im.Exts)
	nf := len(fls)
	if nf == 0 {
		err := fmt.Errorf("Images.OpenNames() -- no image files in: %s", im.Path)
		log.Println(err)
		return err
	}
	sort.Strings(fls)
	im.ImagesAll = make([][]string, 0)
	curcat := ""
	si := 0
	for ni, nm := range fls {
		cat := im.Cat(nm)
		if cat != curcat {
			if curcat != "" {
				im.Cats = append(im.Cats, curcat)
				im.ImagesAll = append(im.ImagesAll, fls[si:ni])
			}
			curcat = cat
			si = ni
		}
	}
	im.Cats = append(im.Cats, curcat)
	im.ImagesAll = append(im.ImagesAll, fls[si:len(fls)])
	im.Split()
	return nil
}

// Split does the train / test split
func (im *Images) Split() {
	if im.SplitByItm {
		im.SplitItems()
	} else {
		im.SplitNoItems()
	}
}

// SplitItems does the train / test split, by items
func (im *Images) SplitItems() {
	nc := len(im.ImagesAll)
	im.ImagesTrain = make([][]string, nc)
	im.ImagesTest = make([][]string, nc)
	for ci, fls := range im.ImagesAll {
		itmp := make(map[string]int)
		for _, f := range fls {
			itm := im.Item(f)
			itmp[itm] = 0
		}
		nitm := len(itmp)
		itms := make([]string, nitm)
		i := 0
		for it := range itmp {
			itms[i] = it
			i++
		}
		pi := rand.Perm(nitm)
		ntst := im.NTestPerCat
		if ntst >= nitm {
			ntst = nitm / 2
		}
		ntrn := nitm - ntst
		tstm := make(map[string]int, ntrn)
		for i = 0; i < ntst; i++ {
			tstm[itms[pi[i]]] = i
		}
		for _, f := range fls {
			itm := im.Item(f)
			_, istst := tstm[itm]
			if istst {
				im.ImagesTest[ci] = append(im.ImagesTest[ci], f)
			} else {
				im.ImagesTrain[ci] = append(im.ImagesTrain[ci], f)
			}
		}
	}
	im.Flats()
}

// SplitNoItems does the train / test split, no items
func (im *Images) SplitNoItems() {
	nc := len(im.ImagesAll)
	im.ImagesTrain = make([][]string, nc)
	im.ImagesTest = make([][]string, nc)
	for ci, fls := range im.ImagesAll {
		nitm := len(fls)
		ntst := im.NTestPerCat
		if ntst >= nitm {
			ntst = nitm / 2
		}
		ntrn := nitm - ntst
		slist := rand.Perm(nitm)
		for i := 0; i < ntrn; i++ {
			im.ImagesTrain[ci] = append(im.ImagesTrain[ci], fls[slist[i]])
		}
		for i := ntrn; i < nitm; i++ {
			im.ImagesTest[ci] = append(im.ImagesTest[ci], fls[slist[i]])
		}
	}
	im.Flats()
}

// SelectCats filters the list of images to those within given list of categories.
func (im *Images) SelectCats(cats []string) {
	nc := len(im.Cats)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := im.Cats[ci]

		sel := false
		for _, cs := range cats {
			if cat == cs {
				sel = true
				break
			}
		}
		if !sel {
			im.Cats = append(im.Cats[:ci], im.Cats[ci+1:]...)
			im.ImagesAll = append(im.ImagesAll[:ci], im.ImagesAll[ci+1:]...)
			im.ImagesTrain = append(im.ImagesTrain[:ci], im.ImagesTrain[ci+1:]...)
			im.ImagesTest = append(im.ImagesTest[:ci], im.ImagesTest[ci+1:]...)
		}
	}
	im.Flats()
}

// DeleteCats filters the list of images to exclude those within given list of categories.
func (im *Images) DeleteCats(cats []string) {
	nc := len(im.Cats)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := im.Cats[ci]

		del := false
		for _, cs := range cats {
			if cat == cs {
				del = true
				break
			}
		}
		if del {
			im.Cats = append(im.Cats[:ci], im.Cats[ci+1:]...)
			im.ImagesAll = append(im.ImagesAll[:ci], im.ImagesAll[ci+1:]...)
			im.ImagesTrain = append(im.ImagesTrain[:ci], im.ImagesTrain[ci+1:]...)
			im.ImagesTest = append(im.ImagesTest[:ci], im.ImagesTest[ci+1:]...)
		}
	}
	im.Flats()
}

// SelectImages filters the list of images to those within given list of images (contains)
func (im *Images) SelectImages(images []string) {
	for ci, _ := range im.ImagesAll {
		ofcat := im.ImagesAll[ci]
		no := len(ofcat)
		for oi := no - 1; oi >= 0; oi-- {
			ofl := ofcat[oi]
			sel := false
			for _, cs := range images {
				if strings.Contains(ofl, cs) {
					sel = true
					break
				}
			}
			if !sel {
				ofcat = append(ofcat[:oi], ofcat[oi+1:]...)
			}
		}
	}
	im.Split()
	im.Flats()
}

// DeleteImages filters the list of images to exclude those within given list of images (contains)
func (im *Images) DeleteImages(images []string) {
	for ci, _ := range im.ImagesAll {
		ofcat := im.ImagesAll[ci]
		no := len(ofcat)
		for oi := no - 1; oi >= 0; oi-- {
			ofl := ofcat[oi]
			del := false
			for _, cs := range images {
				if strings.Contains(ofl, cs) {
					del = true
					break
				}
			}
			if del {
				ofcat = append(ofcat[:oi], ofcat[oi+1:]...)
			}
		}
	}
	im.Split()
	im.Flats()
}

// Flats generates flat lists from categorized lists, in form categ/fname.obj
func (im *Images) Flats() {
	im.FlatAll = im.FlatImpl(im.ImagesAll)
	im.FlatTrain = im.FlatImpl(im.ImagesTrain)
	im.FlatTest = im.FlatImpl(im.ImagesTest)
}

// FlatImpl generates flat lists from categorized lists, in form categ/fname.obj
func (im *Images) FlatImpl(images [][]string) []string {
	var flat []string
	for ci, fls := range images {
		cat := im.Cats[ci]
		for _, fn := range fls {
			if im.CatSep == "" {
				fn = cat + "/" + fn
			}
			flat = append(flat, fn)
		}
	}
	return flat
}
