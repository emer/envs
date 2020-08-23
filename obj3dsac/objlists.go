// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// this file contains lists of objects used in different scenarios

var (
	// ObjsBigSlow are particularly big and slow object files -- avoid if possible -- use in
	// Delete
	ObjsBigSlow = []string{"drums", "tank", "domestictree", "chandelier", "bicycle", "bow"}

	// Objs20 is set of 20 objects used in: O’Reilly, R. C., Russin, J. L., Zolfaghar, M., & Rohrlich, J. (2020). Deep Predictive Learning in Neocortex and Pulvinar. ArXiv:2006.14800 [q-Bio]. http://arxiv.org/abs/2006.14800
	Objs20 = []string{"banana", "car", "chair", "donut", "doorknob", "elephant", "fish", "guitar", "handgun", "heavycannon", "layercake", "motorcycle", "person", "piano", "sailboat", "slrcamera", "stapler", "tablelamp", "trafficcone", "trex"}
)
