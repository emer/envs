// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

var AllConditions = map[string]*Condition{
	"PosAcq_B50": {
		Name:      "PosAcq_B50",
		Desc:      "Standard positive valence acquisition: A = 100%, B = 50%",
		Block:     "PosAcq_B50",
		FixedProb: true,
		NBlocks:   51,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcq_A50": {
		Name:      "PosAcq_A50",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos at 50%",
		Block:     "PosAcq_A50",
		FixedProb: true,
		NBlocks:   51,
		NTrials:   10,
		Permute:   true,
	},
	"US0": {
		Name:      "US0",
		Desc:      "No US at all",
		Block:     "US0",
		FixedProb: true,
		NBlocks:   5,
		NTrials:   100,
		Permute:   true,
	},
	"PosAcqPreSecondOrder": {
		Name:      "PosAcqPreSecondOrder",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, B at 50%",
		Block:     "PosAcqPreSecondOrder",
		FixedProb: true,
		NBlocks:   51,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcq_B50Cont": {
		Name:      "PosAcq_B50Cont",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, B at 50% reinf, tags further learning as reacq",
		Block:     "PosReacq",
		FixedProb: true,
		NBlocks:   50,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcq_B100": {
		Name:      "PosAcq_B100",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, B at 100%",
		Block:     "PosAcq_B100",
		FixedProb: true,
		NBlocks:   50,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcq_B100Cont": {
		Name:      "PosAcq_B100Cont",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos -- continue w/ wts",
		Block:     "PosAcq_B100",
		FixedProb: true,
		NBlocks:   50,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcqEarlyUS_test": {
		Name:      "PosAcqEarlyUS_test",
		Desc:      "Testing session: after pos_acq trng, deliver US early or late",
		Block:     "PosAcqEarlyUS_test",
		FixedProb: true,
		NBlocks:   5,
		NTrials:   2,
		Permute:   false,
	},
	"PosAcq_B25": {
		Name:      "PosAcq_B25",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos",
		Block:     "PosAcq_B25",
		FixedProb: true,
		NBlocks:   200,
		NTrials:   8,
		Permute:   true,
	},
	"PosExtinct": {
		Name:      "PosExtinct",
		Desc:      "Pavlovian extinction: A_NR_Pos",
		Block:     "PosExtinct",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   8,
		Permute:   true,
	},
	"PosCondInhib": {
		Name:      "PosCondInhib",
		Desc:      "conditioned inhibition training: AX_NR_Pos, A_R_Pos interleaved",
		Block:     "PosCondInhib",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   8,
		Permute:   true,
	},
	"PosSecondOrderCond": {
		Name:      "PosSecondOrderCond",
		Desc:      "second order conditioning training: AB_NR_Pos, A_R_Pos interleaved; A = 1st order, F = 2nd order CS",
		Block:     "PosSecondOrderCond",
		FixedProb: false,
		NBlocks:   10,
		NTrials:   50,
		Permute:   true,
	},
	"PosCondInhib_test": {
		Name:      "PosCondInhib_test",
		Desc:      "Testing session: A_NR_Pos, AX_NR_Pos, and X_NR_Pos cases",
		Block:     "PosCondInhib_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   6,
		Permute:   false,
	},
	"NegAcq": {
		Name:      "NegAcq",
		Desc:      "Pavlovian conditioning w/ negatively-valenced US: D_R_NEG",
		Block:     "NegAcq",
		FixedProb: false,
		NBlocks:   76,
		NTrials:   10,
		Permute:   true,
	},
	"NegAcqFixedProb": {
		Name:      "NegAcqFixedProb",
		Desc:      "Pavlovian conditioning w/ negatively-valenced US: A_R_NEG",
		Block:     "NegAcq",
		FixedProb: true,
		NBlocks:   150,
		NTrials:   8,
		Permute:   false,
	},
	"PosAcqOmit": {
		Name:      "PosAcqOmit",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, A_NR_Pos trials, interleaved",
		Block:     "PosAcqOmit",
		FixedProb: false,
		NBlocks:   10,
		NTrials:   8,
		Permute:   true,
	},
	"NegCondInh": {
		Name:      "NegCondInh",
		Desc:      "condition inhibition w/ negatively-valenced US: CZ_NR_NEG, C_R_NEG interleaved; i.e.,  Z = security signal",
		Block:     "NegCondInhib",
		FixedProb: false,
		NBlocks:   75,
		NTrials:   10,
		Permute:   true,
	},
	"NegCondInh_test": {
		Name:      "NegCondInh_test",
		Desc:      "condition inhibition w/ negatively-valenced US: CZ_NR_NEG, C_R_NEG interleaved; i.e.,  Z = security signal",
		Block:     "NegCondInhib_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   6,
		Permute:   false,
	},
	"NegExtinct": {
		Name:      "NegExtinct",
		Desc:      "Pavlovian conditioning w/ negatively-valenced US: A_R_NEG",
		Block:     "NegExtinct",
		FixedProb: false,
		NBlocks:   75,
		NTrials:   8,
		Permute:   true,
	},
	"PosAcq_cxA": {
		Name:      "PosAcq_cxA",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, A_R_Pos_omit trials, interleaved",
		Block:     "PosAcq_cxA",
		FixedProb: false,
		NBlocks:   26,
		NTrials:   10,
		Permute:   false,
	},
	"PosExtinct_cxB": {
		Name:      "PosExtinct_cxB",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, A_R_Pos_omit trials, interleaved",
		Block:     "PosExtinct_cxB",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   10,
		Permute:   false,
	},
	"PosRenewal_cxA": {
		Name:      "PosRenewal_cxA",
		Desc:      "Pavlovian conditioning w/ positively-valenced US: A_R_Pos, A_R_Pos_omit trials, interleaved",
		Block:     "PosRenewal_cxA",
		FixedProb: false,
		NBlocks:   1,
		NTrials:   2,
		Permute:   false,
	},
	"PosBlocking_A_train": {
		Name:      "PosBlocking_A_train",
		Desc:      "Blocking experiment",
		Block:     "PosBlocking_A_train",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   1,
		Permute:   false,
	},
	"PosBlocking": {
		Name:      "PosBlocking",
		Desc:      "Blocking experiment",
		Block:     "PosBlocking",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   2,
		Permute:   false,
	},
	"PosBlocking_test": {
		Name:      "PosBlocking_test",
		Desc:      "Blocking experiment",
		Block:     "PosBlocking_test",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   1,
		Permute:   false,
	},
	"PosBlocking2_test": {
		Name:      "PosBlocking2_test",
		Desc:      "Blocking experiment",
		Block:     "PosBlocking2_test",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   2,
		Permute:   false,
	},
	"NegBlocking_E_train": {
		Name:      "NegBlocking_E_train",
		Desc:      "Blocking experiment",
		Block:     "NegBlocking_E_train",
		FixedProb: false,
		NBlocks:   300,
		NTrials:   1,
		Permute:   false,
	},
	"NegBlocking": {
		Name:      "NegBlocking",
		Desc:      "Blocking experiment",
		Block:     "NegBlocking",
		FixedProb: false,
		NBlocks:   200,
		NTrials:   2,
		Permute:   false,
	},
	"NegBlocking_test": {
		Name:      "NegBlocking_test",
		Desc:      "Blocking experiment",
		Block:     "NegBlocking_test",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   1,
		Permute:   false,
	},
	"PosAcqMag": {
		Name:      "PosAcqMag",
		Desc:      "Magnitude experiment",
		Block:     "PosAcqMagnitude",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   8,
		Permute:   false,
	},
	"PosSumAcq": {
		Name:      "PosSumAcq",
		Desc:      "Conditioned Inhibition - A+, C+",
		Block:     "PosSumAcq",
		FixedProb: false,
		NBlocks:   450,
		NTrials:   3,
		Permute:   false,
	},
	"PosSumCondInhib": {
		Name:      "PosSumCondInhib",
		Desc:      "Conditioned Inhibition - AX-, A+",
		Block:     "PosCondInhib_BY",
		FixedProb: false,
		NBlocks:   300,
		NTrials:   3,
		Permute:   false,
	},
	"PosSum_test": {
		Name:      "PosSum_test",
		Desc:      "Conditioned Inhibition Summation Test",
		Block:     "PosSumCondInhib_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   6,
		Permute:   false,
	},
	"NegSumAcq": {
		Name:      "NegSumAcq",
		Desc:      "Conditioned Inhibition - D-, E-",
		Block:     "NegSumAcq",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   3,
		Permute:   false,
	},
	"NegSumCondInhib": {
		Name:      "NegSumCondInhib",
		Desc:      "Conditioned Inhibition - DU, D-",
		Block:     "NegCondInhib_FV",
		FixedProb: false,
		NBlocks:   100,
		NTrials:   3,
		Permute:   false,
	},
	"NegSum_test": {
		Name:      "NegSum_test",
		Desc:      "Conditioned Inhibition Summation Test",
		Block:     "NegSumCondInhib_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   6,
		Permute:   false,
	},
	"Unblocking_train": {
		Name:      "Unblocking_train",
		Desc:      "A+++,B+++,C+",
		Block:     "Unblocking_train",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   2,
		Permute:   false,
	},
	"UnblockingValue": {
		Name:      "UnblockingValue",
		Desc:      "AX+++,CZ+++",
		Block:     "UnblockingValue",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   1,
		Permute:   false,
	},
	"UnblockingValue_test": {
		Name:      "UnblockingValue_test",
		Desc:      "A,X,C,Z",
		Block:     "UnblockingValue_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   1,
		Permute:   false,
	},
	"Unblocking_trainUS": {
		Name:      "Unblocking_trainUS",
		Desc:      "A+++ (water) ,B+++ (food)",
		Block:     "Unblocking_trainUS",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   15,
		Permute:   false,
	},
	"UnblockingIdentity": {
		Name:      "UnblockingIdentity",
		Desc:      "AX+++(water),BY+++(water)",
		Block:     "UnblockingIdentity",
		FixedProb: false,
		NBlocks:   25,
		NTrials:   20,
		Permute:   false,
	},
	"UnblockingIdentity_test": {
		Name:      "UnblockingIdentity_test",
		Desc:      "A,X,B,Y",
		Block:     "UnblockingIdentity_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   4,
		Permute:   false,
	},
	"PosAcqMagChange": {
		Name:      "PosAcqMagChange",
		Desc:      "Magnitude experiment",
		Block:     "PosAcqMagnitudeChange",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   4,
		Permute:   false,
	},
	"NegAcqMag": {
		Name:      "NegAcqMag",
		Desc:      "Magnitude experiment",
		Block:     "NegAcqMagnitude",
		FixedProb: false,
		NBlocks:   51,
		NTrials:   8,
		Permute:   false,
	},
	"NegAcqMagChange": {
		Name:      "NegAcqMagChange",
		Desc:      "Magnitude experiment",
		Block:     "NegAcqMagnitudeChange",
		FixedProb: false,
		NBlocks:   50,
		NTrials:   4,
		Permute:   false,
	},
	"Overexpect_train": {
		Name:      "Overexpect_train",
		Desc:      "Overexpectation training (A+, B+, C+, X+, Y-)",
		Block:     "Overexpectation_train",
		FixedProb: false,
		NBlocks:   150,
		NTrials:   5,
		Permute:   false,
	},
	"OverexpectCompound": {
		Name:      "OverexpectCompound",
		Desc:      "Overexpectation compound training (AX+, BY-, CX+, X+, Y-)",
		Block:     "OverexpectationCompound",
		FixedProb: false,
		NBlocks:   150,
		NTrials:   5,
		Permute:   false,
	},
	"Overexpect_test": {
		Name:      "Overexpect_test",
		Desc:      "Overexpectation test ( A-, B-, C-, X-)",
		Block:     "Overexpectation_test",
		FixedProb: false,
		NBlocks:   5,
		NTrials:   5,
		Permute:   false,
	},
	"PosNeg": {
		Name:      "PosNeg",
		Desc:      "Positive negative test - W equally reinforced with reward + punishment",
		Block:     "PosNeg",
		FixedProb: false,
		NBlocks:   150,
		NTrials:   6,
		Permute:   false,
	},
	"PosOrNegAcq": {
		Name:      "PosOrNegAcq",
		Desc:      "Positive negative acquisition - with reward or punishment on interleaved trials according to user-set probabilities",
		Block:     "PosOrNegAcq",
		FixedProb: false,
		NBlocks:   150,
		NTrials:   6,
		Permute:   true,
	},
	"USDebug": {
		Name:      "USDebug",
		Desc:      "For debugging, 100% reward, CS A",
		Block:     "USDebug",
		FixedProb: true,
		NBlocks:   51,
		NTrials:   8,
		Permute:   true,
	},
}
