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
	"os"
	"strconv"
	"testing"
)

/*

go test -bench=IRQDetails -run=^$ -cpu=1,4 -benchmem -benchtime=10s

goos: linux
goarch: amd64
pkg: github.com/thediveo/irks
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkIRQDetailsOsReadDir                8337           1278516 ns/op          632801 B/op       1255 allocs/op
BenchmarkIRQDetailsOsReadDir-4              8763           1340999 ns/op          633533 B/op       1255 allocs/op
BenchmarkIRQDetails                        15948            739545 ns/op           65787 B/op        413 allocs/op
BenchmarkIRQDetails-4                      16454            726045 ns/op           65796 B/op        413 allocs/op

*/

// AllIRQDetailsTraditional does it the traditional Gopher way, using
// os.File.ReadDir and os.ReadFile, so we have the benchmarking reference.
func AllIRQDetailsTraditional(root string) iter.Seq[IRQDetails] {
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

			contents, err := os.ReadFile(root + syskernelirqPath + irqEntry.Name() + actionsNode)
			if err != nil || len(contents) < 1 || contents[len(contents)-1] != '\n' {
				continue
			}
			details.Actions = string(contents[:len(contents)-1])

			contents, err = os.ReadFile(root + procirqPath + irqEntry.Name() + effectiveAffinityNode)
			if err != nil || len(contents) < 1 || contents[len(contents)-1] != '\n' {
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

func BenchmarkIRQDetailsOsReadDir(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range AllIRQDetailsTraditional("") {
		}
	}
}

func BenchmarkIRQDetails(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range allIRQDetails("") {
		}
	}
}
