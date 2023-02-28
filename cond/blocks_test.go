// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cond

import (
	"fmt"
	"strings"
	"testing"
)

func TestCSContext(t *testing.T) {
	css := make(map[string]int)
	for blnm, bl := range AllBlocks {
		for _, trl := range bl {
			cnt := css[trl.CS]
			css[trl.CS] = cnt + 1

			if trl.CS == "" {
				t.Errorf("CS is empty: %s   in block: %s  trial: %s\n", trl.CS, blnm, trl.Name)
			}
			if trl.Context == "" {
				fmt.Printf("Context: %s empty -- will be copied to CS: %s   in block: %s  trial: %s\n", trl.Context, trl.CS, blnm, trl.Name)
			}
			if trl.Context != trl.CS {
				fmt.Printf("Context: %s != CS: %s   in block: %s  trial: %s\n", trl.Context, trl.CS, blnm, trl.Name)
			}
			if len(trl.CS) > 1 && trl.CS2Start <= 0 {
				t.Errorf("CS has multiple elements but CS2Start is not set: %s   in block: %s  trial: %s\n", trl.CS, blnm, trl.Name)
			}
			if trl.CS2Start > 0 {
				if len(trl.CS) != 2 {
					t.Errorf("CS2Start is set but CS != 2 elements: %s   in block: %s  trial: %s\n", trl.CS, blnm, trl.Name)
				}
				// fmt.Printf("CS2Start: %d  CS: %s   in block: %s  trial: %s\n", trl.CS2Start, trl.CS, blnm, trl.Name)
			}
			if strings.Contains(trl.Name, "_R") && trl.USProb == 0 {
				fmt.Printf("_R trial with USProb = 0 in block: %s  trial: %s\n", blnm, trl.Name)
			}
			if strings.Contains(trl.Name, "_NR") && trl.USProb != 0 {
				fmt.Printf("_NR trial with USProb != 0 in block: %s  trial: %s\n", blnm, trl.Name)
			}
			if strings.Contains(trl.Name, "_test") && !trl.Test {
				fmt.Printf("_test Trial.Name with Test = false in block: %s  trial: %s\n", blnm, trl.Name)
			}
			if strings.Contains(blnm, "_test") && !trl.Test {
				fmt.Printf("_test Block name with Test = false in block: %s  trial: %s\n", blnm, trl.Name)
			}
		}
	}
	fmt.Printf("\nList of unique CSs and use count:\n")
	for cs, cnt := range css {
		fmt.Printf("%s \t %d\n", cs, cnt)
	}
}

func TestConds(t *testing.T) {
	for cnm, cd := range AllConditions {
		_, ok := AllBlocks[cd.Block]
		if !ok {
			t.Errorf("Block name: %s not found in Condition: %s\n", cd.Block, cnm)
		}
	}
}
