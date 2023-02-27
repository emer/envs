// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

// Condition defines parameters for running a specific type of conditioning expt
type Condition struct {
	Name          string `desc:"identifier for this type of configuration"`
	Desc          string `desc:"description of this configuration"`
	Block         string `desc:"type of block to run"`
	FixedProb     bool   `desc:"fixed probability for each trial group"`
	NIters        int    `desc:"number of iterations to run"`
	BlocksPerIter int    `desc:"number of blocks (1 block = one behavioral trial = sequence of CS, US) in each iteration -- needs to be higher if there are stochastic variables (probabilities)."`
	Permute       bool   `desc:"permute list of fully-instantiated trials after generation"`
	LoadWeights   bool   `desc:"load initial weights from a file (specified in weights_file)"`
	WeightsFile   string `desc:"full relative path (from project) of weights file to load -- use CRR: prefix to load from cluster run results directory"`
	LoadStBlk     int    `desc:"after loading weights, reset block counter to this value (-1 = leave at value from the loaded weights)"`
}
