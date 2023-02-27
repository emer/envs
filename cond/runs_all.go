// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

var AllRuns = map[string]Run{
	"RunMaster": {
		Name:  "RunMaster",
		Desc:  "",
		Cond1: "PosAcq_B50",
	},
	"USDebug": {
		Name:  "USDebug",
		Desc:  "",
		Cond1: "USDebug",
	},
	"US0": {
		Name:  "US0",
		Desc:  "",
		Cond1: "US0",
	},
	"PosAcq_A50": {
		Name:  "PosAcq_A50",
		Desc:  "",
		Cond1: "PosAcq_A50",
	},
	"PosAcq_B50Ext": {
		Name:  "PosAcq_B50Ext",
		Desc:  "",
		Cond1: "PosAcq_B50",
		Cond2: "PosExtinct",
	},
	"PosAcq_B50ExtAcq": {
		Name:  "PosAcq_B50ExtAcq",
		Desc:  "Full cycle: acq, ext, acq",
		Cond1: "PosAcq_B50",
		Cond2: "PosExtinct",
		Cond3: "PosAcq_B50Cont",
	},
	"PosAcq_B100Ext": {
		Name:  "PosAcq_B100Ext",
		Desc:  "",
		Cond1: "PosAcq_B100",
		Cond2: "PosExtinct",
	},
	"PosAcq": {
		Name:  "PosAcq",
		Desc:  "",
		Cond1: "PosAcq_B50",
	},
	"PosExt": {
		Name:  "PosExt",
		Desc:  "",
		Cond1: "PosAcq_B50",
		Cond2: "PosExtinct",
	},
	"PosAcq_B25": {
		Name:  "PosAcq_B25",
		Desc:  "",
		Cond1: "PosAcq_B25",
	},
	"NegAcq": {
		Name:  "NegAcq",
		Desc:  "",
		Cond1: "NegAcq",
	},
	"NegAcqMag": {
		Name:  "NegAcqMag",
		Desc:  "",
		Cond1: "NegAcqMag",
	},
	"PosAcqMag": {
		Name:  "PosAcqMag",
		Desc:  "",
		Cond1: "PosAcqMag",
	},
	"NegAcqExt": {
		Name:  "NegAcqExt",
		Desc:  "",
		Cond1: "NegAcq",
		Cond2: "NegExtinct",
	},
	"PosCondInhib": {
		Name:  "PosCondInhib",
		Desc:  "",
		Cond1: "PosAcq_contextA",
		Cond2: "PosCondInhib",
		Cond3: "PosCondInhib_test",
	},
	"PosSecondOrderCond": {
		Name:  "PosSecondOrderCond",
		Desc:  "",
		Cond1: "PosAcqPreSecondOrder",
		Cond2: "PosSecondOrderCond",
	},
	"PosBlocking": {
		Name:  "PosBlocking",
		Desc:  "",
		Cond1: "PosBlocking_A_training",
		Cond2: "PosBlocking",
		Cond3: "PosBlocking_test",
	},
	"PosBlocking2": {
		Name:  "PosBlocking2",
		Desc:  "",
		Cond1: "PosBlocking_A_training",
		Cond2: "PosBlocking",
		Cond3: "PosBlocking2_test",
	},
	"NegCondInhib": {
		Name:  "NegCondInhib",
		Desc:  "",
		Cond1: "NegAcq",
		Cond2: "NegCondInh",
		Cond3: "NegCondInh_test",
	},
	"AbaRenewal": {
		Name:  "AbaRenewal",
		Desc:  "",
		Cond1: "PosAcq_contextA",
		Cond2: "PosExtinct_contextB",
		Cond3: "PosRenewal_contextA",
	},
	"NegBlocking": {
		Name:  "NegBlocking",
		Desc:  "",
		Cond1: "NegBlocking_E_training",
		Cond2: "NegBlocking",
		Cond3: "NegBlocking_test",
	},
	"PosSum_test": {
		Name:  "PosSum_test",
		Desc:  "",
		Cond1: "PosSumAcq",
		Cond2: "PosSumCondInhib",
		Cond3: "PosSum_test",
	},
	"NegSum_test": {
		Name:  "NegSum_test",
		Desc:  "",
		Cond1: "NegSumAcq",
		Cond2: "NegSumCondInhib",
		Cond3: "NegSum_test",
	},
	"UnblockingValue": {
		Name:  "UnblockingValue",
		Desc:  "",
		Cond1: "Unblocking_train",
		Cond2: "UnblockingValue",
		Cond3: "UnblockingValue_test",
	},
	"UnblockingIdentity": {
		Name:  "UnblockingIdentity",
		Desc:  "",
		Cond1: "Unblocking_trainUS",
		Cond2: "UnblockingIdentity",
		Cond3: "UnblockingIdentity_test",
	},
	"Overexpect": {
		Name:  "Overexpect",
		Desc:  "",
		Cond1: "Overexpect_train",
		Cond2: "OverexpectCompound",
		Cond3: "Overexpect_test",
	},
	"PosMagChange": {
		Name:  "PosMagChange",
		Desc:  "",
		Cond1: "PosAcqMag",
		Cond2: "PosAcqMagChange",
		Cond3: "Overexpect_test",
	},
	"NegMagChange": {
		Name:  "NegMagChange",
		Desc:  "",
		Cond1: "NegAcqMag",
		Cond2: "NegAcqMagChange",
	},
	"CondExp": {
		Name:  "CondExp",
		Desc:  "",
		Cond1: "CondExp",
	},
	"PainExp": {
		Name:  "PainExp",
		Desc:  "",
		Cond1: "PainExp",
	},
	"PosNeg": {
		Name:  "PosNeg",
		Desc:  "",
		Cond1: "PosOrNegAcq",
	},
	"PosAcqEarlyUSTest": {
		Name:  "PosAcqEarlyUSTest",
		Desc:  "",
		Cond1: "PosAcq_B50",
		Cond2: "PosAcqEarlyUS_test",
	},
	"PosOrNegAcq": {
		Name:  "PosOrNegAcq",
		Desc:  "",
		Cond1: "PosOrNegAcq",
	},
	"PosCondInhib_test": {
		Name:  "PosCondInhib_test",
		Desc:  "For debugging",
		Cond1: "PosCondInhib_test",
	},
}
