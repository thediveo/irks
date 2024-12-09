// Copyright 2024 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package irks

import (
	"iter"

	"github.com/thediveo/faf"
)

// IRQDetails provides the list of actions and the currently set CPU affinities
// for a specific IRQ, as indicated by Num.
type IRQDetails struct {
	Num        uint          // IRQ number
	Actions    string        // list of IRQ actions
	Affinities CPUAffinities // effective CPU(s) affinities
}

// CPUAffinities is a list of CPU [from...to] ranges. CPU numbers are starting
// from zero.
type CPUAffinities [][2]uint

// AllIRQDetails returns an iterator looping over the details of all
// (non-architecture-specific) IRQs in the system, giving their details as to
// actions and CPU affinities.
//
// AllIRQDetails uses a streamlined implementation that runs at approx 1.9× the
// execution speed compared to a “traditional” Go implementation approach using
// [os.File.ReadDir], [strconv.ParseUint] and [os.ReadFile]. For the same system
// with 47 hardware IRQs, we only need at around 10% of memory on the heap, and
// only 1/3 of allocations, compared to using stock stdlib functions.
func AllIRQDetails() iter.Seq[IRQDetails] {
	return allIRQDetails("")
}

const (
	syskernelirqPath = "/sys/kernel/irq/"
	procirqPath      = "/proc/irq/"

	actionsNode           = "/actions"
	effectiveAffinityNode = "/effective_affinity_list"
)

func allIRQDetails(root string) iter.Seq[IRQDetails] {
	return func(yield func(IRQDetails) bool) {
		// Using bytes.Buffer instead of assembling path strings piecewise
		// doesn't buy us anything above the noise floor, even with
		// preallocating the buffer's capacity once and then truncating back to
		// the root. But reusing the buffer to read the pseudo files boosts us...
		var contents []byte
		var details IRQDetails
		for irqEntry := range faf.ReadDir(root + syskernelirqPath) {
			if !irqEntry.IsDir() {
				continue
			}
			irqnum, ok := faf.ParseUint(irqEntry.Name)
			if !ok {
				continue
			}
			details.Num = uint(irqnum)

			contents, ok := faf.ReadFile(
				root+syskernelirqPath+string(irqEntry.Name)+actionsNode, contents)
			if !ok || len(contents) < 1 || contents[len(contents)-1] != '\n' {
				continue
			}
			details.Actions = string(contents[:len(contents)-1]) // escapes

			contents, ok = faf.ReadFile(
				root+procirqPath+string(irqEntry.Name)+effectiveAffinityNode, contents)
			if !ok || len(contents) < 1 || contents[len(contents)-1] != '\n' {
				continue
			}
			afflist := cpuList(contents[:len(contents)-1])
			if len(afflist) == 0 {
				continue
			}
			details.Affinities = afflist

			if !yield(details) {
				return
			}
		}
	}
}
