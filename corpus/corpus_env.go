// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/evec"
	"github.com/emer/emergent/patgen"
	"github.com/emer/empi/empi"
	"github.com/emer/empi/mpi"
	"github.com/emer/etable/etensor"
)

// CorpusEnv reads text from a file and presents a window of sequential words
// as the input.  Words included in the vocabulary can be filtered by frequency
// at both ends.
// For input, a simple bag-of-words is used, with words encoded as localist 1-hot
// units (requiring a large input layer) or using random distributed vectors in a
// lower-dimensional space.
type CorpusEnv struct {

	// name of this environment
	Nm string `desc:"name of this environment"`

	// description of this environment
	Dsc string `desc:"description of this environment"`

	// full list of words used for activating state units according to index
	Words []string `desc:"full list of words used for activating state units according to index"`

	// map of words onto index in Words list
	WordMap map[string]int `desc:"map of words onto index in Words list"`

	// map of words onto frequency in entire corpus, normalized
	FreqMap map[string]float64 `desc:"map of words onto frequency in entire corpus, normalized"`

	// entire corpus as one long list of words
	Corpus []string `desc:"entire corpus as one long list of words"`

	// full list of sentences
	Sentences [][]string `desc:"full list of sentences"`

	// offsets into corpus for each sentence
	SentOffs []int `desc:"offsets into corpus for each sentence"`

	// map of words into random distributed vector encodings
	WordReps etensor.Float32 `desc:"map of words into random distributed vector encodings"`

	// list of words in the current window
	InputWords []string `desc:"list of words in the current window"`

	// current window activation state
	Input etensor.Float32 `desc:"current window activation state"`

	// instead of presenting full sentences, this env just probes each word in turn
	ProbeMode bool `desc:"instead of presenting full sentences, this env just probes each word in turn"`

	// size of sliding window of words to show in input
	WindowSize int `desc:"size of sliding window of words to show in input"`

	// use localist 1-hot encoding of words -- else random dist vectors
	Localist bool `desc:"use localist 1-hot encoding of words -- else random dist vectors"`

	// distributed representations of words: target percent activity total for WindowSize words all on at same time
	DistPctAct float64 `desc:"distributed representations of words: target percent activity total for WindowSize words all on at same time"`

	// randomly drop out the highest-frequency inputs
	DropOut bool `desc:"randomly drop out the highest-frequency inputs"`

	// use an UNK token for unknown words -- otherwise just skip
	UseUNK bool `desc:"use an UNK token for unknown words -- otherwise just skip"`

	// size of input layer state
	InputSize evec.Vec2i `desc:"size of input layer state"`

	// maximum number of words representable -- InputSize.X*Y
	MaxVocab int `desc:"maximum number of words representable -- InputSize.X*Y"`

	// location of the generated vocabulary file
	VocabFile string `desc:"location of the generated vocabulary file"`

	// for this processor (MPI), starting index into Corpus
	CorpStart int `inactive:"+" desc:"for this processor (MPI), starting index into Corpus"`

	// for this processor (MPI), ending index into Corpus
	CorpEnd int `inactive:"+" desc:"for this processor (MPI), ending index into Corpus"`

	// [view: inline] current run of model as provided during Init
	Run env.Ctr `view:"inline" desc:"current run of model as provided during Init"`

	// [view: inline] epoch is arbitrary increment of number of times through trial.Max steps
	Epoch env.Ctr `view:"inline" desc:"epoch is arbitrary increment of number of times through trial.Max steps"`

	// [view: inline] trial is the network training step counter
	Trial env.Ctr `view:"inline" desc:"trial is the network training step counter"`

	// [view: inline] tick counts steps through the Corpus
	Tick env.Ctr `view:"inline" desc:"tick counts steps through the Corpus"`

	// [view: inline] block counts iterations through the entire Corpus
	Block env.Ctr `view:"inline" desc:"block counts iterations through the entire Corpus"`
}

var epsilon = 1e-7

func (ev *CorpusEnv) Name() string { return ev.Nm }
func (ev *CorpusEnv) Desc() string { return ev.Dsc }

// InitTMat initializes matrix and labels to given size
func (ev *CorpusEnv) Validate() error {
	return nil
}

func (ev *CorpusEnv) State(element string) etensor.Tensor {
	switch element {
	case "Input":
		return &ev.Input
	}
	return nil
}

func (ev *CorpusEnv) Init(run int) {
	ev.Run.Scale = env.Run
	ev.Epoch.Scale = env.Epoch
	ev.Trial.Scale = env.Trial
	ev.Run.Init()
	ev.Epoch.Init()
	ev.Trial.Init()
	ev.Run.Cur = run
	ev.Trial.Cur = -1 // init state -- key so that first Step() = 0
}

func (ev *CorpusEnv) Config(probe bool, inputfile string, windowsize int, inputsize evec.Vec2i, localist, dropout bool, distPctAct float64) {
	ev.ProbeMode = probe
	ev.WindowSize = windowsize
	ev.InputSize = inputsize
	ev.Localist = localist
	ev.DistPctAct = distPctAct
	ev.MaxVocab = ev.InputSize.X * ev.InputSize.Y
	ev.DropOut = dropout
	ev.UseUNK = false
	ev.VocabFile = "stored_vocab_cbt_train.json"

	ev.LoadFmFile(inputfile)
	ev.SentToCorpus()

	ev.Input.SetShape([]int{ev.InputSize.Y, ev.InputSize.X}, nil, []string{"Y", "X"})
	ev.InputWords = make([]string, ev.WindowSize)

	if !ev.Localist {
		ev.ConfigWordReps()
	}

	if ev.ProbeMode {
		nwords := len(ev.Words)
		cmodn := nwords / mpi.WorldSize()
		cmodn *= mpi.WorldSize() // actual number we can process under mpi
		ev.CorpStart, ev.CorpEnd, _ = empi.AllocN(cmodn)
	} else {
		corpn := len(ev.Corpus)
		cmodn := corpn / mpi.WorldSize()
		cmodn *= mpi.WorldSize() // actual number we can process under mpi
		ev.CorpStart, ev.CorpEnd, _ = empi.AllocN(cmodn)
	}
	ev.Tick.Max = ev.CorpEnd - ev.CorpStart
}

// JsonData is the full original corpus, in sentence form
type JsonData struct {
	Vocab     []string
	Freqs     []int
	Sentences [][]string
}

// JsonVocab is the map-based encoding of just the vocabulary and frequency
type JsonVocab struct {
	Vocab map[string]int
	Freqs map[string]float64
}

func (ev *CorpusEnv) NormalizeFreqs() {
	var s float64
	s = 0
	for w, _ := range ev.FreqMap {
		// ev.FreqMap[w] = math.Sqrt(f)
		s += ev.FreqMap[w]
	}
	for w, _ := range ev.FreqMap {
		ev.FreqMap[w] /= s
	}
}

func (ev *CorpusEnv) LoadFmFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	var data JsonData
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return err
	}
	ev.Words = data.Vocab
	ev.Sentences = data.Sentences
	ev.SentOffs = make([]int, len(ev.Sentences)+1)
	ev.SentOffs[0] = 0
	for i := 1; i <= len(ev.Sentences); i++ {
		ev.SentOffs[i] = ev.SentOffs[i-1] + len(ev.Sentences[i-1])
	}

	ev.WordMap = make(map[string]int)
	ev.FreqMap = make(map[string]float64)
	_, err = os.Stat(ev.VocabFile)
	if os.IsNotExist(err) {
		fmt.Println("Building new vocabulary")

		if ev.UseUNK {
			ev.WordMap["[UNK]"] = 0
			ev.FreqMap["[UNK]"] = 0
		}
		for i, w := range ev.Words {
			ev.WordMap[w] = len(ev.WordMap)
			ev.FreqMap[w] = float64(data.Freqs[i])
		}
		ev.NormalizeFreqs()

		if len(ev.WordMap) > ev.MaxVocab {
			err := fmt.Errorf("Vocab size read %d out of bounds.", len(ev.WordMap))
			fmt.Println(err.Error())
			return err
		}

		var jobj JsonVocab
		jobj.Vocab = ev.WordMap
		jobj.Freqs = ev.FreqMap
		jenc, _ := json.MarshalIndent(jobj, "", " ")
		_ = ioutil.WriteFile(ev.VocabFile, jenc, 0644)

	} else {
		fmt.Println("Loading existing vocabulary")

		vocab, err := os.Open(ev.VocabFile)
		if err != nil {
			fmt.Println("Error opening vocabulary:", err)
			return err
		}
		defer vocab.Close()

		var jobj JsonVocab
		byteValue, _ := ioutil.ReadAll(vocab)
		if err := json.Unmarshal(byteValue, &jobj); err != nil {
			fmt.Println(err)
			return err
		}
		ev.WordMap = jobj.Vocab
		ev.FreqMap = jobj.Freqs
	}

	return nil
}

// SentToCorpus makes the Corpus out of the Sentences
func (ev *CorpusEnv) SentToCorpus() {
	ev.Corpus = make([]string, 0, len(ev.SentOffs)-1)
	for si := 0; si < len(ev.Sentences); si++ {
		for wi := 0; wi < len(ev.Sentences[si]); wi++ {
			w := ev.LookUpWord(ev.Sentences[si][wi])
			if w != "" {
				ev.Corpus = append(ev.Corpus, w)
			}
		}
	}
}

func (ev *CorpusEnv) ConfigWordReps() {
	nwords := len(ev.Words)
	nin := ev.InputSize.X * ev.InputSize.Y
	nun := int(ev.DistPctAct * float64(nin))
	nper := nun / ev.WindowSize // each word has this many active, assuming no overlap
	mindif := nper / 2

	ev.WordReps.SetShape([]int{nwords, ev.InputSize.Y, ev.InputSize.X}, nil, []string{"Y", "X"})

	fname := fmt.Sprintf("word_reps_%dx%d_on%d_mind%d.json", ev.InputSize.Y, ev.InputSize.X, nper, mindif)

	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		fmt.Printf("ConfigWordReps: nwords: %d  nin: %d  nper: %d  minDif: %d\n", nwords, nin, nper, mindif)

		patgen.MinDiffPrintIters = true
		patgen.PermutedBinaryMinDiff(&ev.WordReps, nper, 1, 0, mindif)
		jenc, _ := json.Marshal(ev.WordReps.Values)
		_ = ioutil.WriteFile(fname, jenc, 0644)
	} else {
		fmt.Printf("Loading word reps from: %s\n", fname)
		file, err := os.Open(fname)
		if err != nil {
			fmt.Println("Error opening file:", err)
		} else {
			defer file.Close()
			bs, _ := ioutil.ReadAll(file)
			if err := json.Unmarshal(bs, &ev.WordReps.Values); err != nil {
				fmt.Println(err)
			}
		}
	}
}

// CorpusPosToSentIdx returns the sentence, idx of given corpus position
func (ev *CorpusEnv) CorpusPosToSentIdx(pos int) []int {
	idx := make([]int, 2)
	var i, j, curr int
	i = 0
	j = len(ev.SentOffs) - 1
	for {
		curr = (i + j) / 2
		if pos < ev.SentOffs[curr] {
			j = curr - 1
		} else {
			i = curr + 1
		}
		if i >= j {
			idx[0] = curr
			idx[1] = pos - ev.SentOffs[idx[0]]
			break
		}
	}
	return idx
}

// String returns the current state as a string
func (ev *CorpusEnv) String() string {
	curr := make([]string, len(ev.InputWords))
	if ev.InputWords != nil {
		copy(curr, ev.InputWords)
		for i := 0; i < len(curr); i++ {
			if curr[i] == "" {
				curr[i] = "--"
			}
		}
		return strings.Join(curr, " ")
	}
	return ""
}

// RenderWords renders the current list of InputWords to Input state
func (ev *CorpusEnv) RenderWords() {
	ev.Input.SetZeros()
	for i := 0; i < ev.WindowSize; i++ {
		if ev.InputWords[i] == "" {
			continue
		}
		widx := ev.WordMap[ev.InputWords[i]]
		if ev.Localist {
			ev.Input.SetFloat1D(widx, 1)
		} else {
			wp := ev.WordReps.SubSpace([]int{widx})
			idx := 0
			for y := 0; y < ev.InputSize.Y; y++ {
				for x := 0; x < ev.InputSize.X; x++ {
					wv := wp.FloatVal1D(idx)
					cv := ev.Input.FloatVal1D(idx)
					nv := math.Max(wv, cv)
					ev.Input.SetFloat1D(idx, nv)
					idx++
				}
			}
		}
	}
}

func (ev *CorpusEnv) LookUpWord(word string) string {
	ans := word
	if _, ok := ev.WordMap[ans]; !ok {
		if ev.UseUNK {
			ans = "[UNK]"
		} else {
			ans = ""
		}
	} else if ev.DropOut {
		samp := rand.Float64()
		if ev.FreqMap[word]-samp > epsilon {
			ans = ""
		}
	}
	return ans
}

// RenderState renders the current state
func (ev *CorpusEnv) RenderState() {
	cur := ev.CorpStart + ev.Tick.Cur
	if ev.ProbeMode {
		ev.InputWords[0] = ev.Words[cur]
		for i := 1; i < ev.WindowSize; i++ {
			ev.InputWords[i] = ""
		}
	} else {
		for i := 0; i < ev.WindowSize; i++ {
			ev.InputWords[i] = ev.Corpus[(cur+i)%len(ev.Corpus)]
		}
	}
	ev.RenderWords()
}

func (ev *CorpusEnv) Step() bool {
	ev.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	ev.Block.Same() // good idea to just reset all non-inner-most counters at start
	if ev.Tick.Incr() {
		ev.Block.Incr()
	}
	if ev.Trial.Incr() {
		ev.Epoch.Incr()
	}
	ev.RenderState()
	return true
}

func (ev *CorpusEnv) Action(element string, input etensor.Tensor) {
	// nop
}

func (ev *CorpusEnv) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return ev.Run.Query()
	case env.Epoch:
		return ev.Epoch.Query()
	case env.Trial:
		return ev.Trial.Query()
	}
	return -1, -1, false
}

// Compile-time check that implements Env interface
var _ env.Env = (*CorpusEnv)(nil)
