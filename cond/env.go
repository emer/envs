// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

import (
	"math/rand"
	"strings"

	"github.com/emer/emergent/erand"
	"github.com/goki/mat32"
)

// GenerateTrials generates specific trial types for given condition name.
// gets the block name from the condition name.
func GenerateTrials(condNm string) []*Trial {
	var trls []*Trial
	cond := AllConditions[condNm]
	block := AllBlocks[cond.Block]
	for _, trl := range block {
		nRpt := int(mat32.Round(trl.Pct * float32(cond.BlocksPerIter)))
		if nRpt < 1 {
			if trl.Pct > 0.0 {
				nRpt = 1
			} else {
				continue // shouldn't happen
			}
		}
		useIsOnList := false
		var usIsOn []bool
		if cond.FixedProb || strings.Contains(trl.Name, "AutoTstEnv") {
			if trl.USProb != 0.0 && trl.USProb != 1.0 {
				useIsOnList = true
				pn := int(mat32.Round(float32(nRpt) * trl.USProb))
				usIsOn = make([]bool, nRpt)
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
			usFlag := false
			if !strings.Contains(trlNm, "NR") { // nonreinforced (NR) trials NEVER get reinforcement
				if !useIsOnList {
					usFlag = erand.BoolP(trl.USProb)
				} else {
					usFlag = usIsOn[ri]
				}
				trlNm = strings.TrimSuffix(trlNm, "_omit")
				if !usFlag {
					trlNm += "_omit"
				}
			}
			testFlag := false
			if strings.Contains(trlNm, "_test") {
				parts := strings.Split(trlNm, "_")
				trlNm = ""
				for _, part := range parts {
					if part != "test" {
						trlNm += part + "_"
					}
				}
				trlNm += "test"
				testFlag = true
			}
			curTrial := &Trial{}
			*curTrial = *trl
			curTrial.Name = trlNm
			curTrial.TestFlag = testFlag
			curTrial.USFlag = usFlag
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

/*
func (ev *PVLVEnv) SetupOneAlphaTrial(curTrial *data.TrialInstance, stimNum int) {
	prefixUSTimeIn := ""

	// CAUTION! - using percent normalization assumes the multiple CSs (e.g., AX) are always on together,
	// i.e., the same timesteps; thus, doesn't work for second-order conditioning
	stimInBase := pvlv.StmNone
	stimIn2Base := pvlv.StmNone
	nStims := 1
	nUSTimes := 1

	// CAUTION! below string-pruning requires particular convention for naming trial_gps in terms of CSs used;
	// e.g., "AX_*", etc.

	// CAUTION! For either multiple CSs (e.g., AX) or mixed_US case
	// (e.g., sometimes reward, sometimes punishment) only two (2) simultaneous representations
	// currently supported; AND, multiple CSs and mixed_US cases can only be used separately, not together;
	// code will need re-write if more complicated cases are desired (e.g., more than two (2) representations
	// or using multiple CSs/mixed_US together).

	cs := curTrial.TrialName[0:2]
	cs1 := ""
	cs2 := ""
	if strings.Contains(cs, "_") {
		cs1 = cs[0:1]
		cs = pvlv.StmNone.String()
		cs2 = pvlv.StmNone.String()
		nStims = 1
		// need one for each predictive CS; also, one for each PREDICTED US if same CS (e.g., Z')
		// predicts two different USs probalistically (i.e., mixed_US == true condition)
		nUSTimes = 1
		stimInBase = pvlv.StimMap[cs1]
	} else {
		cs1 = cs[0:1]
		cs2 = cs[1:2]
		nStims = 2
		// need one for each predictive CS; also, one for each PREDICTED US if same CS (e.g., Z')
		// predicts two different USs probalistically (i.e., mixed_US == true condition)
		nUSTimes = 2
		stimInBase = pvlv.StimMap[cs1]
		stimIn2Base = pvlv.StimMap[cs2]
	}

	// Set up Context_In reps

	// initialize to use the basic context_in var to rep the basic case in which CS and Context are isomorphic
	ctxParts := pvlv.CtxRe.FindStringSubmatch(curTrial.Context)
	ctx1 := ctxParts[1]
	ctx2 := ctxParts[2]
	preContext := ctx1 + ctx2
	postContext := ctxParts[3]
	contextIn := pvlv.CtxMap[curTrial.Context]
	contextIn2 := pvlv.CtxNone
	contextIn3 := pvlv.CtxNone
	nContexts := len(preContext)
	// gets complicated if more than one CS...
	if len(preContext) > 1 {
		switch ev.ContextModel {
		case ELEMENTAL:
			// first element, e.g., A
			contextIn = pvlv.CtxMap[ctx1]
			// second element, e.g., X
			contextIn2 = pvlv.CtxMap[ctx2]
			// only handles two for now...
		case CONJUNCTIVE:
			// use "as is"...
			contextIn = pvlv.CtxMap[curTrial.Context]
			nContexts = 1
		case BOTH:
			// first element, e.g., A
			contextIn = pvlv.CtxMap[ctx1]
			// second element, e.g., X
			contextIn2 = pvlv.CtxMap[ctx2]
			// conjunctive case, e.g., AX
			contextIn3 = pvlv.CtxMap[preContext]
			nContexts = len(preContext) + 1
		}
	}
	// anything after the "_" indicates different context for extinction, renewal, etc.
	if len(postContext) > 0 {
		contextIn = pvlv.CtxMap[ctx1+"_"+postContext]
		if len(ctx2) > 0 {
			contextIn2 = pvlv.CtxMap[ctx2+"_"+postContext]
		}
		contextIn3 = pvlv.CtxNone
	}

	if ev.StdInputData.Rows != 0 {
		ev.StdInputData.SetNumRows(0)
	}

	// configure and write all the leabra trials for one eco trial
	for i := 0; i < curTrial.AlphaTicksPerTrialGp; i++ {
		i := ev.AlphaCycle.Cur
		alphaTrialName := curTrial.TrialName + "_t" + strconv.Itoa(i)
		trialGpTimestep := pvlv.Tick(i)
		trialGpTimestepInt := i
		stimIn := pvlv.StmNone
		stimIn2 := pvlv.StmNone
		posPV := pvlv.PosUSNone
		negPV := pvlv.NegUSNone
		usTimeInStr := ""
		usTimeIn2Str := ""
		usTimeInWrongStr := ""
		usTimeIn := pvlv.USTimeNone
		usTimeIn2 := pvlv.USTimeNone
		usTimeInWrong := pvlv.USTimeNone
		notUSTimeIn := pvlv.USTimeNone
		prefixUSTimeIn = cs1 + "_"
		prefixUSTimeIn2 := ""
		if nUSTimes == 2 {
			prefixUSTimeIn2 = cs2 + "_"
		}
		// set CS input activation values on or off according to timesteps
		// set first CS - may be the only one
		if i >= curTrial.CSTimeStart && i <= curTrial.CSTimeEnd {
			stimIn = stimInBase
			// TODO: Theoretically, USTime reps shouldn't come on at CS-onset until BAacq and/or
			// gets active first - for time being, using a priori inputs as a temporary proof-of-concept
		} else {
			stimIn = pvlv.StmNone
		}
		// set CS2 input activation values on or off according to timesteps, if a second CS exists
		if i >= curTrial.CS2TimeStart && i <= curTrial.CS2TimeEnd {
			stimIn2 = stimIn2Base
		} else {
			stimIn2 = pvlv.StmNone
		}
		// set US and USTime input activation values on or off according to timesteps
		var us int
		if i > curTrial.CSTimeStart && (!(i > curTrial.USTimeStart) || !curTrial.USFlag) {
			if curTrial.ValenceContext == pvlv.POS {
				us = int(pvlv.PosSMap[curTrial.USType])
				posPV = pvlv.PosUS(us)
				usTimeInStr = prefixUSTimeIn + "PosUS" + strconv.Itoa(us) + "_t" +
					strconv.Itoa(i-curTrial.CSTimeStart-1)
				usTimeIn = pvlv.PUSTFromString(usTimeInStr)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				if curTrial.MixedUS {
					usTimeInWrongStr = prefixUSTimeIn + "NegUS" + strconv.Itoa(us) + "_t" +
						strconv.Itoa(i-curTrial.CSTimeStart-1)
					usTimeInWrong = pvlv.PUSTFromString(usTimeInWrongStr)
				}
			} else if curTrial.ValenceContext == pvlv.NEG {
				us = int(pvlv.NegSMap[curTrial.USType])
				negPV = pvlv.NegUS(us)
				usTimeInStr = prefixUSTimeIn + "NegUS" + strconv.Itoa(us) + "_t" +
					strconv.Itoa(i-curTrial.CSTimeStart-1)
				usTimeIn = pvlv.PUSTFromString(usTimeInStr)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				if curTrial.MixedUS {
					usTimeInWrongStr = prefixUSTimeIn + "PosUS" + strconv.Itoa(us) + "_t" +
						strconv.Itoa(i-curTrial.CSTimeStart-1)
					usTimeInWrong = pvlv.PUSTFromString(usTimeInWrongStr)
				}
			}
		} else {
			usTimeIn = pvlv.USTimeNone
			notUSTimeIn = pvlv.USTimeNone
			usTimeInStr = pvlv.USTimeNone.String()
		}

		if i > curTrial.CS2TimeStart && i <= (curTrial.CS2TimeEnd+1) && (!(i > curTrial.USTimeStart) || !curTrial.USFlag) {
			usTime2IntStr := strconv.Itoa(i - curTrial.CS2TimeStart - 1)
			if curTrial.ValenceContext == pvlv.POS {
				us = int(pvlv.PosSMap[curTrial.USType])
				posPV = pvlv.PosUS(us)
				usTimeIn2Str = prefixUSTimeIn2 + "PosUS" + strconv.Itoa(us) + "_t" + usTime2IntStr
				usTimeIn2 = pvlv.PUSTFromString(usTimeIn2Str)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				if curTrial.MixedUS {
					usTimeInWrongStr = prefixUSTimeIn + "NegUS" + strconv.Itoa(us) + "_t" + usTime2IntStr
					usTimeInWrong = pvlv.USTimeNone.FromString(usTimeInWrongStr)
				}
			} else if curTrial.ValenceContext == pvlv.NEG {
				negPV = pvlv.NegSMap[curTrial.USType]
				us = int(negPV)
				usTimeIn2Str = prefixUSTimeIn2 + "NegUS" + strconv.Itoa(us) + "_t" + usTime2IntStr
				usTimeIn2 = pvlv.PUSTFromString(usTime2IntStr)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				if curTrial.MixedUS {
					usTimeInWrongStr = prefixUSTimeIn + "PosUS" + strconv.Itoa(us) + "_t" +
						strconv.Itoa(i-curTrial.CSTimeStart-1)
					usTimeInWrong = pvlv.USTimeNone.FromString(usTimeInWrongStr)
				}
			}
		} else {
			usTimeIn2 = pvlv.USTimeNone
			notUSTimeIn = pvlv.USTimeNone
			usTimeIn2Str = pvlv.USTimeNone.String()
		}

		if (i >= curTrial.USTimeStart) && (i <= curTrial.USTimeEnd) && curTrial.USFlag {
		} else {
			posPV = pvlv.PosUSNone
			negPV = pvlv.NegUSNone
		}
		if (i > curTrial.USTimeStart) && curTrial.USFlag {
			if curTrial.ValenceContext == pvlv.POS {
				us = int(pvlv.PosSMap[curTrial.USType])
				usTimeInStr = "PosUS" + strconv.Itoa(us) + "_t" + strconv.Itoa(i-curTrial.USTimeStart-1)
				usTimeIn = pvlv.USTimeNone.FromString(usTimeInStr)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				usTimeInWrong = pvlv.USTimeNone
			} else if curTrial.ValenceContext == pvlv.NEG {
				us = int(pvlv.NegSMap[curTrial.USType])
				usTimeInStr = "NegUS" + strconv.Itoa(us) + "_t" + strconv.Itoa(i-curTrial.USTimeStart-1)
				usTimeIn = pvlv.USTimeNone.FromString(usTimeInStr)
				usTimeInWrongStr = pvlv.USTimeNone.String()
				usTimeInWrong = pvlv.USTimeNone
			}
		}
		pvEmpty := pvlv.PosUSNone.Tensor()
		curTimestepStr := ""
		curTimeStepInt := 0
		stimulus := ""
		stimDenom := 1.0
		ctxtDenom := 1.0
		usTimeDenom := 1.0
		curTimeStepInt = trialGpTimestepInt
		curTimestepStr = trialGpTimestep.String()
		if stimNum == 0 {
			ev.StdInputData.AddRows(1)
		}
		if nStims == 1 {
			stimulus = stimIn.String()
		} else {
			stimulus = cs1 + cs2
		} // // i.e., there is a 2nd stimulus, e.g., 'AX', 'BY'

		ev.StdInputData.SetCellString("AlphTrialName", curTimeStepInt, alphaTrialName)
		ev.StdInputData.SetCellString("Time", curTimeStepInt, curTimestepStr)
		ev.StdInputData.SetCellString("Stimulus", curTimeStepInt, stimulus)
		ev.StdInputData.SetCellString("Context", curTimeStepInt, curTrial.Context)

		tsrStim := etensor.NewFloat64(pvlv.StimInShape, nil, nil)
		tsrCtx := etensor.NewFloat64(pvlv.ContextInShape, nil, nil)
		if curTimeStepInt >= curTrial.CSTimeStart && curTimeStepInt <= curTrial.CSTimeEnd {
			stimDenom = 1.0 + ev.PctNormTotalActStim*float64(nStims-1)
			if stimIn != pvlv.StmNone {
				tsrStim.SetFloat([]int{int(stimIn)}, 1.0/stimDenom)
			}
			if stimIn2 != pvlv.StmNone {
				tsrStim.SetFloat([]int{int(stimIn2)}, 1.0/stimDenom)
			}
			ev.StdInputData.SetCellTensor("StimIn", curTimeStepInt, tsrStim)

			ctxtDenom = 1.0 + ev.PctNormTotalActCtx*float64(nContexts-1)
			if contextIn != pvlv.CtxNone {
				tsrCtx.SetFloat(contextIn.Parts(), 1.0/ctxtDenom)
			}
			if contextIn3 != pvlv.CtxNone {
				tsrCtx.SetFloat(contextIn3.Parts(), 1.0/ctxtDenom)
			}
			ev.StdInputData.SetCellTensor("ContextIn", curTimeStepInt, tsrCtx)
		}
		if curTimeStepInt >= curTrial.CS2TimeStart && curTimeStepInt <= curTrial.CS2TimeEnd {
			stimDenom = 1.0 + ev.PctNormTotalActStim*float64(nStims-1)
			if stimIn2 != pvlv.StmNone {
				tsrStim.SetFloat([]int{int(stimIn2)}, 1.0/stimDenom)
			}
			ev.StdInputData.SetCellTensor("StimIn", curTimeStepInt, tsrStim)

			ctxtDenom = 1.0 + ev.PctNormTotalActCtx*float64(nContexts-1)
			if contextIn2 != pvlv.CtxNone {
				tsrCtx.SetFloat(contextIn2.Parts(), 1.0/ctxtDenom)
			}
			if contextIn3 != pvlv.CtxNone {
				tsrCtx.SetFloat(contextIn3.Parts(), 1.0/ctxtDenom)
			}
			ev.StdInputData.SetCellTensor("ContextIn", curTimeStepInt, tsrCtx)
		}

		if curTrial.USFlag && (curTimeStepInt >= curTrial.USTimeStart && curTimeStepInt <= curTrial.USTimeEnd) {
			if curTrial.USFlag && curTrial.ValenceContext == pvlv.POS {
				if posPV != pvlv.PosUSNone {
					ev.StdInputData.SetCellTensor("PosPV", curTimeStepInt, posPV.Tensor())
				} else {
					ev.StdInputData.SetCellTensor("PosPV", curTimeStepInt, pvEmpty)
				}
			} else if curTrial.USFlag && curTrial.ValenceContext == pvlv.NEG {
				if negPV != pvlv.NegUSNone {
					ev.StdInputData.SetCellTensor("NegPV", curTimeStepInt, negPV.Tensor())
				} else {
					ev.StdInputData.SetCellTensor("NegPV", curTimeStepInt, pvEmpty)
				}
			}
		} else {
			ev.StdInputData.SetCellTensor("PosPV", curTimeStepInt, pvEmpty)
			ev.StdInputData.SetCellTensor("NegPV", curTimeStepInt, pvEmpty)
		}

		usTimeDenom = 1.0 + ev.PctNormTotalActUSTime*float64(nUSTimes-1)
		tsrUSTime := etensor.NewFloat64(pvlv.USTimeInShape, nil, nil)
		if usTimeIn != pvlv.USTimeNone {
			setVal := usTimeIn.Unpack().Coords()
			tsrUSTime.SetFloat(setVal, 1.0/usTimeDenom)
		}
		if usTimeIn2 != pvlv.USTimeNone {
			setVal := usTimeIn2.Unpack().Coords()
			tsrUSTime.SetFloat(setVal, 1.0/usTimeDenom)
		}
		if usTimeInWrong != pvlv.USTimeNone {
			tsrUSTime.SetFloat(usTimeInWrong.Shape(), 1.0/usTimeDenom)
		}
		if notUSTimeIn != pvlv.USTimeNone {
			tsrUSTime.SetFloat(notUSTimeIn.Shape(), 1.0/usTimeDenom)
		}
		ev.StdInputData.SetCellTensor("USTimeIn", curTimeStepInt, tsrUSTime)
		if usTimeIn2Str != "" {
			usTimeIn2Str = "+" + usTimeIn2Str + usTimeIn2.Unpack().CoordsString()
		}
		ev.StdInputData.SetCellString("USTimeInStr", curTimeStepInt,
			usTimeInStr+usTimeIn.Unpack().CoordsString()+usTimeIn2Str)
	}
}

*/
