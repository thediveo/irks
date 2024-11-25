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
	"math/rand/v2"
	"os"
	"regexp"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

func collectIRQs(it iter.Seq[IRQ]) []IRQ {
	irqs := []IRQ{}
	for irq := range it {
		irq := irq
		irq.Counters = slices.Clone(irq.Counters)
		irqs = append(irqs, irq)
	}
	return irqs
}

const procInterruptsText = ` CPU1 CPU42 CPU666
 1: 2 3 4 x
 5: 6 7 8 y
 ENEMIH: 1 2 3 zz
`

var _ = Describe("irksome", func() {

	When("determining online CPU numbers", func() {

		It("returns an empty list for malformed lines", func() {
			Expect(cpuListFromProcInterrupts([]byte(""))).To(BeEmpty())
			Expect(cpuListFromProcInterrupts([]byte("  FOO0 FOO1"))).To(BeEmpty())
			Expect(cpuListFromProcInterrupts([]byte("  CPUA CPU42"))).To(BeEmpty())
		})

		It("returns the correct list", func() {
			Expect(cpuListFromProcInterrupts([]byte("  CPU1  CPU42  CPU666 "))).To(
				HaveExactElements(CPUList{1, 42, 666}))
		})

	})

	When("reading all IRQ counters", func() {

		It("yields nothing for invalid data", func() {
			r := strings.NewReader("")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader("\n")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n ")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n 1")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n 1: 2")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n 1: 2 ")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n 1: 2 abc")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())

			r = strings.NewReader(" CPU1 CPU2\n 1: 2abc 3")
			Expect(collectIRQs(allCounters(r, nil))).To(BeEmpty())
		})

		It("yields the correct IRQ information", func() {
			r := strings.NewReader(` CPU1 CPU42 CPU666
 1: 2 3 4 x
 5: 6 7 8 y
`)
			irqs := collectIRQs(allCounters(r, nil))
			Expect(irqs).To(HaveEach(
				HaveField("CPUs", HaveExactElements(uint(1), uint(42), uint(666)))))
			Expect(irqs).To(HaveExactElements(
				And(HaveField("Num", uint(1)),
					HaveField("Counters", HaveExactElements(uint64(2), uint64(3), uint64(4)))),
				And(HaveField("Num", uint(5)),
					HaveField("Counters", HaveExactElements(uint64(6), uint64(7), uint64(8))))))

			r = strings.NewReader(procInterruptsText)
			irqs = collectIRQs(allCounters(r, nil))
			Expect(irqs).To(HaveEach(
				HaveField("CPUs", HaveExactElements(uint(1), uint(42), uint(666)))))
			Expect(irqs).To(HaveExactElements(
				And(HaveField("Num", uint(1)),
					HaveField("Counters", HaveExactElements(uint64(2), uint64(3), uint64(4)))),
				And(HaveField("Num", uint(5)),
					HaveField("Counters", HaveExactElements(uint64(6), uint64(7), uint64(8))))))
		})

		It("stops the yield when told", func() {
			r := strings.NewReader(procInterruptsText)
			items := 0
			for range allCounters(r, nil) {
				items++
				break
			}
			Expect(items).To(Equal(1))
		})

		It("reads something sensible from /proc/interrupts", func() {
			procinterrupts := Successful(os.ReadFile("/proc/interrupts"))
			numIRQs := len(regexp.MustCompile(`(?m)^\s*\d+:.+`).FindAllString(string(procinterrupts), -1))
			Expect(numIRQs).NotTo(BeZero())
			Expect(AllCounters()).To(HaveLen(numIRQs))
			numCPUs := 0
			for irq := range AllCounters() {
				if numCPUs == 0 {
					numCPUs = len(irq.CPUs)
				}
				Expect(irq.Counters).To(HaveLen(numCPUs))
			}
			Expect(numIRQs).NotTo(BeZero())
		})

	})

	When("wanting only counters for certain IRQs", func() {

		It("yields the correct IRQ information", func() {
			r := strings.NewReader(` CPU1 CPU42 CPU666
 1: 2 3 4 x
 42: 6 7 8 y
 666: 9 10 11 z
 888: 21 22 23 abc
`)
			irqs := collectIRQs(allCounters(r, []uint{1, 666}))
			Expect(irqs).To(HaveLen(2))
			Expect(irqs).To(HaveExactElements(
				HaveField("Num", uint(1)),
				HaveField("Num", uint(666))))
		})

		It("produces only wanted IRQ information", func() {
			allirqs := collectIRQs(AllCounters())
			irqnums := []uint{}
			for i := 0; i < 3; i++ {
				var randomirq uint
				for {
					randomirq = allirqs[rand.UintN(uint(len(allirqs)))].Num
					if !slices.Contains(irqnums, randomirq) {
						break
					}
				}
				irqnums = append(irqnums, randomirq)
			}
			slices.Sort(irqnums)
			irqs := collectIRQs(CountersFor(irqnums))
			Expect(irqs).To(HaveLen(3))
			for i, irqnum := range irqnums {
				Expect(irqs[i].Num).To(Equal(irqnum))
			}
		})

	})

	When("getting CPU affinities", func() {

		DescribeTable("parsing CPU lists",
			func(s string, aff CPUAffinities) {
				Expect(cpuList([]byte(s))).To(Equal(aff))
			},
			Entry(nil, "", CPUAffinities{}),
			Entry(nil, "a", CPUAffinities{}),
			Entry(nil, "42-", CPUAffinities{}),
			Entry(nil, "42!", CPUAffinities{}),
			Entry(nil, "42", CPUAffinities{{42, 42}}),
			Entry(nil, "42,666", CPUAffinities{{42, 42}, {666, 666}}),
			Entry(nil, "42-666", CPUAffinities{{42, 666}}),
			Entry(nil, "42,44-45", CPUAffinities{{42, 42}, {44, 45}}),
			Entry(nil, "42,44-45,666", CPUAffinities{{42, 42}, {44, 45}, {666, 666}}),
		)

	})

	When("getting IRQ details", func() {

		It("returns nothing then there are errors", func() {
			Expect(allIRQDetails("./testdata/non-existing")).To(BeEmpty())

		})

		It("returns correct details", func() {
			Expect(allIRQDetails("./testdata/mixed")).To(ConsistOf(
				IRQDetails{
					Num:        42,
					Actions:    []string{"foo", "bar"},
					Affinities: CPUAffinities{{1, 3}, {42, 42}},
				},
				IRQDetails{
					Num:        43,
					Actions:    []string{"baz"},
					Affinities: CPUAffinities{{0, 8}, {15, 15}},
				}))
		})

		It("aborts iterator", func() {
			counts := 0
			for _ = range allIRQDetails("./testdata/mixed") {
				counts++
				break
			}
			Expect(counts).To(Equal(1))
		})

		It("reads real IRQ details", func() {
			counts := 0
			irqnums := map[uint]struct{}{}
			for irq := range AllCounters() {
				irqnums[irq.Num] = struct{}{}
			}
			for irqdetail := range AllIRQDetails() {
				counts++
				Expect(irqnums).To(HaveKey(irqdetail.Num))
				Expect(irqdetail.Actions).NotTo(BeEmpty())
				Expect(irqdetail.Affinities).NotTo(BeEmpty())
			}
			Expect(counts).NotTo(BeZero())
		})

	})

})
