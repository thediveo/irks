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
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			Expect(cpuList([]byte(""))).To(BeEmpty())
			Expect(cpuList([]byte("  FOO0 FOO1"))).To(BeEmpty())
			Expect(cpuList([]byte("  CPUA CPU42"))).To(BeEmpty())
		})

		It("returns the correct list", func() {
			Expect(cpuList([]byte("  CPU1  CPU42  CPU666 "))).To(
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
			numIRQs := 0
			numCPUs := 0
			for irq := range AllCounters() {
				numIRQs++
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
					randomirq = rand.UintN(uint(len(allirqs)))
					if !slices.Contains(irqnums, randomirq) {
						break
					}
				}
				irqnums = append(irqnums, allirqs[randomirq].Num)
			}
			slices.Sort(irqnums)
			irqs := collectIRQs(CountersFor(irqnums))
			Expect(irqs).To(HaveLen(3))
			for i, irqnum := range irqnums {
				Expect(irqs[i].Num).To(Equal(irqnum))
			}
		})

	})

})
