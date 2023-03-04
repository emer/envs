// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

import (
	"math/rand"

	"github.com/emer/emergent/erand"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

//go:generate stringer -type=Valence

// Valence
type Valence int32

const (
	Pos Valence = iota
	Neg
	ValenceN
)

var KiT_Valence = kit.Enums.AddEnum(ValenceN, kit.NotBitFlag, nil)

func (ev Valence) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Valence) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// Trial represents one behavioral trial, unfolding over
// NTicks individual time steps, with one or more CS's (conditioned stimuli)
// and a US (unconditioned stimulus -- outcome).
type Trial struct {
	Name     string  `desc:"conventional suffixes: _R = reward, _NR = non-reward; _test = test trial (no learning)"`
	Test     bool    `desc:"true if testing only -- no learning"`
	Pct      float32 `desc:"Percent of all trials for this type"`
	Valence  Valence `desc:"Positive or negative reward valence"`
	USProb   float32 `desc:"Probability of US"`
	MixedUS  bool    `desc:"Mixed US set?"`
	USMag    float32 `desc:"US magnitude"`
	NTicks   int     `desc:"Number of ticks for a trial"`
	CS       string  `desc:"Conditioned stimulus"`
	CSStart  int     `desc:"Tick of CS start"`
	CSEnd    int     `desc:"Tick of CS end"`
	CS2Start int     `desc:"Tick of CS2 start: -1 for none"`
	CS2End   int     `desc:"Tick of CS2 end: -1 for none"`
	US       int     `desc:"Unconditioned stimulus"`
	USStart  int     `desc:"Tick for start of US presentation"`
	USEnd    int     `desc:"Tick for end of US presentation"`
	Context  string  `desc:"Context -- typically same as CS -- if blank CS will be copied -- different in certain extinguishing contexts"`

	// The following are set for generated trials:
	USOn bool `desc:"for rendered trials, true if US active"`
	CSOn bool `desc:"for rendered trials, true if CS active"`
}

// Block represents a set of trial types
type Block []*Trial

func (cd *Block) Length() int {
	return len(*cd)
}

func (cd *Block) Append(trl *Trial) {
	*cd = append(*cd, trl)
}

// GenerateTrials generates repetitions of specific trial types
// for given condition name, based on Pct of total blocks,
// and sets the USOn flag for proportion of trials
// based on USProb probability.
// If Condition.Permute is true, order of all trials is permuted.
// Gets the block name from the condition name.
func GenerateTrials(condNm string) []*Trial {
	var trls []*Trial
	cond := AllConditions[condNm]
	block := AllBlocks[cond.Block]
	for _, trl := range block {
		if trl.Context == "" {
			trl.Context = trl.CS
		}
		nRpt := int(mat32.Round(trl.Pct * float32(cond.NTrials)))
		if nRpt < 1 {
			if trl.Pct > 0.0 {
				nRpt = 1
			} else {
				continue // shouldn't happen
			}
		}
		useIsOnList := false
		var usIsOn []bool
		if cond.FixedProb {
			if trl.USProb != 0.0 && trl.USProb != 1.0 {
				useIsOnList = true
				pn := int(mat32.Round(float32(nRpt) * trl.USProb))
				usIsOn = make([]bool, nRpt) // defaults to false
				for i := 0; i < pn; i++ {
					usIsOn[i] = true
				}
				rand.Shuffle(len(usIsOn), func(i, j int) {
					usIsOn[i], usIsOn[j] = usIsOn[j], usIsOn[i]
				})
			}
		}
		for ri := 0; ri < nRpt; ri++ {
			trlNm := trl.Name + "_" + trl.Valence.String()
			usOn := false
			if !useIsOnList {
				usOn = erand.BoolP(trl.USProb)
			} else {
				usOn = usIsOn[ri]
			}
			curTrial := &Trial{}
			*curTrial = *trl
			curTrial.Name = trlNm
			curTrial.USOn = usOn
			trls = append(trls, curTrial)
		}
	}
	if cond.Permute {
		rand.Shuffle(len(trls), func(i, j int) {
			trls[i], trls[j] = trls[j], trls[i]
		})
	}
	return trls
}
