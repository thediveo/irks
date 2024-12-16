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
BenchmarkIRQDetailsOsReadDir                9699           1140566 ns/op          632802 B/op       1255 allocs/op
BenchmarkIRQDetailsOsReadDir-4              9818           1166176 ns/op          633568 B/op       1255 allocs/op
BenchmarkIRQDetails                        17660            675672 ns/op           65787 B/op        413 allocs/op
BenchmarkIRQDetails-4                      18022            662177 ns/op           65795 B/op        413 allocs/op

*/

// AllIRQDetailsOsReadDir does it the traditional Gopher way, using
// os.File.ReadDir and os.ReadFile, so we have the benchmarking reference.
func AllIRQDetailsOsReadDir(root string) iter.Seq[IRQDetails] {
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

// Benchmark reading IRQ details using the traditional best practise including
// os.ReadDir and os.ReadFile.
func BenchmarkIRQDetailsOsReadDir(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range AllIRQDetailsOsReadDir("") {
		}
	}
}

// Benchmark reading IRQ details using optimized faf.ReadDir, faf.ReadFile,
// faf.ParseUint, et cetera.
func BenchmarkIRQDetails(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range allIRQDetails("") {
		}
	}
}
