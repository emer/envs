// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

//go:generate stringer -type=Valence

// Valence
type Valence int32

const (
	Pos Valence = iota
	Neg
	ValenceN
)

// Trial parameterizes a trial type
type Trial struct {
	Name      string  `desc:"name"`
	Pct       float32 `desc:"Percent of all trials for this type"`
	Valence   Valence `desc:"Positive or negative reward valence"`
	USProb    float32 `desc:"Probability of US"`
	USFlag    bool    `desc:"for rendered trials, true if US active"`
	TestFlag  bool    `desc:"for testing trials?"`
	FixedProb bool    `desc:"use a permuted list to ensure an exact number of trials have US -- else random draw each time"`
	MixedUS   bool    `desc:"Mixed US set?"`
	USMag     float32 `desc:"US magnitude"`
	NTicks    int     `desc:"Number of ticks for a trial"`
	CS        string  `desc:"Conditioned stimulus"`
	CSStart   int     `desc:"Tick of CS start"`
	CSEnd     int     `desc:"Tick of CS end"`
	CS2Start  int     `desc:"Tick of CS2 start"`
	CS2End    int     `desc:"Tick of CS2 end"`
	US        int     `desc:"Unconditioned stimulus"`
	USStart   int     `desc:"Tick for start of US presentation"`
	USEnd     int     `desc:"Tick for end of US presentation"`
	Context   string  `desc:"Context"`
}

// Block represents a set of trial types
type Block []*Trial

func (cd *Block) Length() int {
	return len(*cd)
}

func (cd *Block) Append(trl *Trial) {
	*cd = append(*cd, trl)
}
