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
	"bufio"
	"io"
	"iter"
	"os"
	"slices"
)

// IRQActions maps IRQ numbers to their registered “actions”; a single IRQ can
// have multiple actions registered.
type IRQActions map[uint][]string

// IRQ holds the per-CPU interrupt counters for this particular IRQ. Please note
// that the counters are valid only for the duration of the yield call producing
// this IRQ data and will then reused/overwritten afterwards. Code that wishes
// to retain the counters needs to make a copy of them.
type IRQ struct {
	Num      uint     // IRQ number
	Counters []uint64 // per-CPU counters, only valid during callback, then reused.
	CPUs     CPUList  // list of the number of the CPUs that are currently online.
}

// CPUList lists the numbers of the CPUs currently being online.
type CPUList []uint

// AllCounters returns a single-use iterator that loops over “/proc/interrupts”
// producing all (non-architecture-specific) IRQs.
//
// The produced IRQ information contains the per-CPU counters for a particular
// IRQ, but only for CPUs that are currently online.
func AllCounters() iter.Seq[IRQ] {
	f, err := os.Open("/proc/interrupts")
	if err != nil {
		return nothing
	}
	defer f.Close()
	return allCounters(f, nil)
}

// CountersFor returns a single-use iterator that loops over “/proc/interrupts”
// producing only the requested IRQs, skipping non-existing IRQs. The list of
// requested IRQs must be sorted in ascending order, but not in condescending
// order.
//
// The produced IRQ information contains the per-CPU counters for a particular
// IRQ, but only for CPUs that are currently online.
func CountersFor(sortedirqnums []uint) iter.Seq[IRQ] {
	f, err := os.Open("/proc/interrupts")
	if err != nil {
		return nothing
	}
	defer f.Close()
	return allCounters(f, sortedirqnums)
}

// allCounters returns an iterator looping over the IRQs with their per-CPU
// counters based on the information in “/proc/interrupts” format and produced
// by the specified reader.
//
// If optionally a non-nil list of IRQ numbers is passed, then allCounters will
// report only IRQs listed and skip parsing counters and yielding them for IRQs
// not listed.
func allCounters(r io.Reader, irqnums []uint) iter.Seq[IRQ] {
	// Please note that sc.Bytes() returns a slice referencing the scanners
	// internal memory that becomes invalid with advancing to the next
	// line/token.
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		return nothing
	}
	// Processing the first line we learn of the CPUs that are actually online
	// (their numbers).
	cpus := cpuList(sc.Bytes())
	numCPUs := len(cpus)
	if numCPUs == 0 {
		return nothing
	}
	return func(yield func(IRQ) bool) {
		irq := IRQ{
			CPUs:     cpus,
			Counters: make([]uint64, len(cpus)),
		}
		for sc.Scan() {
			// Fetch the IRQ number from the beginning of the current text line,
			// ending the iteration when encountering an "unnumbered"
			// (architecture specific) IRQ.
			bstr := newBytestring(sc.Bytes())
			if bstr.SkipSpace() {
				return
			}
			irqno, ok := bstr.Uint64()
			if !ok {
				return
			}
			if !bstr.SkipText(":") {
				return
			}

			// If IRQ filtering is in place, take heed.
			if irqnums != nil {
				if _, ok := slices.BinarySearch(irqnums, uint(irqno)); !ok {
					continue
				}
			}
			irq.Num = uint(irqno)

			// Now consume the per-CPU counters
			for idx := 0; idx < numCPUs; idx++ {
				if bstr.SkipSpace() {
					return
				}
				count, ok := bstr.Uint64()
				if !ok {
					return
				}
				irq.Counters[idx] = count
			}

			// Push the counters for this IRQ to the consumer of this iterator.
			if !yield(irq) {
				return
			}
		}
	}
}

func nothing(func(IRQ) bool) {}

// cpuList returns the list of CPUs that are currently online, according to the
// passed text line that must be in the format of the header line from
// “/proc/interrupts”.
func cpuList(b []byte) CPUList {
	bstr := newBytestring(b)
	numCPUs := bstr.NumFields()
	if numCPUs == 0 {
		return nil
	}
	cpuNums := make(CPUList, numCPUs)
	idx := 0
	for {
		if bstr.SkipSpace() {
			break
		}
		if !bstr.SkipText("CPU") {
			break
		}
		cpuNum, ok := bstr.Uint64()
		if !ok {
			break
		}
		cpuNums[idx] = uint(cpuNum)
		idx++
	}
	if idx != numCPUs {
		return nil
	}
	return cpuNums
}
