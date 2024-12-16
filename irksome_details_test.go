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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("irksome details", func() {

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
					Actions:    "foo,bar",
					Affinities: CPUAffinities{{1, 3}, {42, 42}},
				},
				IRQDetails{
					Num:        43,
					Actions:    "baz",
					Affinities: CPUAffinities{{0, 8}, {15, 15}},
				}))
		})

		It("aborts iterator", func() {
			counts := 0
			for range allIRQDetails("./testdata/mixed") {
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
