// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

// Run is a sequence of Conditions to run in order
type Run struct {
	Name  string `desc:"Name of the run"`
	Desc  string `desc:"Description"`
	Cond1 string `desc:"name of condition 1"`
	Cond2 string `desc:"name of condition 2"`
	Cond3 string `desc:"name of condition 3"`
	Cond4 string `desc:"name of condition 4"`
	Cond5 string `desc:"name of condition 5"`
}
