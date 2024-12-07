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
	"strconv"
	"strings"
)

// IRQ holds the per-CPU interrupt counters for this particular IRQ. Please note
// that the counters are valid only for the duration of the yield call producing
// this IRQ data and will then reused/overwritten afterwards. Code that wishes
// to retain the counters needs to make a copy of them.
type IRQ struct {
	Num      uint     // IRQ number
	Counters []uint64 // per-CPU counters, valid during a single iteration, then reused.
	CPUs     CPUList  // list of the number of the CPUs that are currently online.
}

// CPUList lists the numbers of the CPUs currently being online. It is used to
// map indices of [IRQ] Counters elements to CPU numbers.
type CPUList []uint

// IRQDetails provides the list of actions and the currently set CPU affinities
// for a specific IRQ, as indicated by Num.
type IRQDetails struct {
	Num        uint          // IRQ number
	Actions    []string      // list of IRQ actions
	Affinities CPUAffinities // effective CPU(s) affinities
}

// CPUAffinities is a list of CPU [from...to] ranges. CPU numbers are starting
// from zero.
type CPUAffinities [][2]uint

// AllCounters returns a single-use iterator that loops over “/proc/interrupts”
// producing all (non-architecture-specific) IRQs.
//
// The produced IRQ information contains the per-CPU counters for a particular
// IRQ, but only for CPUs that are currently online.
func AllCounters() iter.Seq[IRQ] {
	return func(yield func(IRQ) bool) {
		f, err := os.Open("/proc/interrupts")
		if err != nil {
			return
		}
		defer f.Close()
		iterateAllCounters(f, nil, yield)
	}
}

// CountersFor returns a single-use iterator that loops over “/proc/interrupts”
// producing only the requested IRQs, skipping non-existing IRQs. The list of
// requested IRQs must be sorted in ascending order, but not in condescending
// order.
//
// The produced IRQ information contains the per-CPU counters for a particular
// IRQ, but only for CPUs that are currently online.
func CountersFor(sortedirqnums []uint) iter.Seq[IRQ] {
	return func(yield func(IRQ) bool) {
		f, err := os.Open("/proc/interrupts")
		if err != nil {
			return
		}
		defer f.Close()
		iterateAllCounters(f, sortedirqnums, yield)
	}
}

// allCounters returns an iterator looping over the IRQs with their per-CPU
// counters based on the information in “/proc/interrupts” format and produced
// by the specified reader.
//
// If optionally a non-nil list of IRQ numbers is passed, then allCounters will
// report only IRQs listed and skip parsing counters and yielding them for IRQs
// not listed.
func allCounters(r io.Reader, irqnums []uint) iter.Seq[IRQ] {
	return func(yield func(IRQ) bool) {
		iterateAllCounters(r, irqnums, yield)
	}
}

func iterateAllCounters(r io.Reader, irqnums []uint, yield func(IRQ) bool) {
	// Please note that sc.Bytes() returns a slice referencing the scanners
	// internal memory that becomes invalid with advancing to the next
	// line/token.
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		return
	}
	// Processing the first line we learn of the CPUs that are actually online
	// (their numbers).
	cpus := cpuListFromProcInterrupts(sc.Bytes())
	numCPUs := len(cpus)
	if numCPUs == 0 {
		return
	}
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

// cpuListFromProcInterrupts returns the list of CPUs that are currently online,
// according to the passed text line that must be in the format of the header
// line from “/proc/interrupts”.
func cpuListFromProcInterrupts(b []byte) CPUList {
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

// cpuList returns the CPUAffinities list from the given string.
func cpuList(b []byte) CPUAffinities {
	bstr := newBytestring(b)
	// nota bene: not using make(...) saves us somehow 3 allocs overall and
	// decreases memory consumption. compiler optimization??
	cpus := CPUAffinities{}
	for {
		if bstr.EOL() {
			break
		}
		from, ok := bstr.Uint64()
		if !ok {
			break
		}
		if bstr.EOL() {
			cpus = append(cpus, [2]uint{uint(from), uint(from)})
			break
		}
		ch, _ := bstr.Next()
		switch ch {
		case ',':
			cpus = append(cpus, [2]uint{uint(from), uint(from)})
		case '-':
			to, ok := bstr.Uint64()
			if !ok {
				break
			}
			cpus = append(cpus, [2]uint{uint(from), uint(to)})
			ch, ok := bstr.Next()
			if !ok || ch != ',' {
				break
			}
		default:
			break
		}
	}
	return cpus
}

// AllIRQDetails returns an iterator looping over the details of all
// (non-architecture-specific) IRQs in the system, giving their details as to
// actions and CPU affinities.
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
		irqDir, err := os.Open(root + syskernelirqPath)
		if err != nil {
			return
		}
		irqDirEntries, err := irqDir.ReadDir(-1)
		irqDir.Close()
		if err != nil {
			return
		}

		// Using bytes.Buffer instead of assembling path strings piecewise
		// doesn't buy us anything above the noise floor, even with
		// preallocating the buffer's capacity once and then truncating back to
		// the root.
		var contents []byte
		var details IRQDetails
		for _, irqEntry := range irqDirEntries {
			if !irqEntry.IsDir() {
				continue
			}
			irqnum, err := strconv.ParseUint(irqEntry.Name(), 10, 64)
			if err != nil {
				continue
			}
			details.Num = uint(irqnum)

			contents, _ = readFile(root+syskernelirqPath+irqEntry.Name()+actionsNode, contents)
			if len(contents) < 1 || contents[len(contents)-1] != '\n' {
				continue
			}
			details.Actions = strings.Split(string(contents[:len(contents)-1]), ",")

			contents, _ = readFile(root+procirqPath+irqEntry.Name()+effectiveAffinityNode, contents)
			if len(contents) < 1 || contents[len(contents)-1] != '\n' {
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

func readFile(name string, buff []byte) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	size := 512
	data := buff[:0]
	if size > cap(data) {
		data = make([]byte, 0, size)
	}

	for {
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				return data, nil
			}
			return data, err
		}

		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
	}
}
