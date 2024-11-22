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

var _ = Describe("byteline", func() {

	When("checking for EOL", func() {

		It("returns EOL for empty line", func() {
			bstr := newBytestring([]byte{})
			Expect(bstr.EOL()).To(BeTrue())
		})

		It("returns EOL for empty line", func() {
			bstr := newBytestring([]byte("foo"))
			Expect(bstr.EOL()).To(BeFalse())
			bstr.pos += 3
			Expect(bstr.EOL()).To(BeTrue())
		})

	})

	When("skipping space", func() {

		It("reports EOL", func() {
			bstr := newBytestring([]byte("   "))
			Expect(bstr.SkipSpace()).To(BeTrue())
		})

		It("advances past spaces", func() {
			bstr := newBytestring([]byte("   foo"))
			Expect(bstr.SkipSpace()).To(BeFalse())
			Expect(bstr.pos).To(Equal(3))
		})

	})

	When("skipping text", func() {

		It("skips only expected text", func() {
			bstr := newBytestring([]byte("foobar"))
			Expect(bstr.SkipText("foo")).To(BeTrue())
			Expect(bstr.pos).To(Equal(3))
		})

		It("doesn't skip unexpected things", func() {
			bstr := newBytestring([]byte("bar"))
			Expect(bstr.SkipText("baz")).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			Expect(bstr.SkipText("barz")).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			Expect(bstr.SkipText("bar")).To(BeTrue())
			Expect(bstr.pos).To(Equal(3))
		})

	})

	When("parsing numbers", func() {

		It("requires at least one digit", func() {
			bstr := &bytestring{b: []byte("")}
			_, ok := bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			bstr = &bytestring{b: []byte("foo")}
			_, ok = bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			bstr = &bytestring{b: []byte("!!!")}
			_, ok = bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))
		})

		It("returns a correct number", func() {
			bstr := newBytestring([]byte("4"))
			num, ok := bstr.Uint64()
			Expect(ok).To(BeTrue())
			Expect(num).To(Equal(uint64(4)))
			Expect(bstr.pos).To(Equal(1))

			bstr = newBytestring([]byte("7foo"))
			num, ok = bstr.Uint64()
			Expect(ok).To(BeTrue())
			Expect(num).To(Equal(uint64(7)))
			Expect(bstr.pos).To(Equal(1))

			bstr = newBytestring([]byte("1234567890123"))
			num, ok = bstr.Uint64()
			Expect(ok).To(BeTrue())
			Expect(num).To(Equal(uint64(1234567890123)))
			Expect(bstr.pos).To(Equal(13))
		})

	})

	When("counting fields", func() {

		It("returns nothing from nothing", func() {
			bstr := newBytestring([]byte(""))
			Expect(bstr.NumFields()).To(BeZero())

			bstr = newBytestring([]byte(" "))
			Expect(bstr.NumFields()).To(BeZero())
		})

		It("counts correctly", func() {
			bstr := newBytestring([]byte(" F  BAR BAZ"))
			Expect(bstr.NumFields()).To(Equal(3))
			bstr = newBytestring([]byte(" F  BAR BAZ RATZ "))
			Expect(bstr.NumFields()).To(Equal(4))
		})

	})

})
