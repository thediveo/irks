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

import "bytes"

// bytestring provides efficient parsing of text lines in form of byte slices.
//
// We rely on the compiler emitting efficient machine code for determining
// len(b), so we don't need to store it explicitly, but use the length already
// stored as part of the byte slice anyway. Looking at the generated x86-64
// machine code using a tools like the “[Compiler Explorer]” confirms efficient
// code where len(bs.b) results in a single MOVQ instruction. For arm64 “all we
// get” is also a single MOVD instruction.
//
// [Compiler Explorer]: https://godbolt.org/
type bytestring struct {
	b   []byte // line contents
	pos int    // parsing position within the line contents
}

// newBytestring returns a new Bytestring object for parsing the supplied text
// line as a byte slice.
func newBytestring(b []byte) *bytestring {
	return &bytestring{
		pos: 0,
		b:   b,
	}
}

// EOL returns true if the parsing has reached the end of the byte string,
// otherwise false.
func (b *bytestring) EOL() (eol bool) { return b.pos >= len(b.b) }

// SkipSpace skips over any space 0x20 characters until either reaching the
// first non-space character, or EOF. When reaching EOL, it returns true.
func (b *bytestring) SkipSpace() (eol bool) {
	for {
		if b.pos >= len(b.b) {
			return true
		}
		if b.b[b.pos] != ' ' {
			return false
		}
		b.pos++
	}
}

// SkipText skips the text s in the buffer at the current position if present,
// returning ok true. Otherwise, returns ok false and the buffer's parsing
// position is left unchanged.
func (b *bytestring) SkipText(s string) (ok bool) {
	if b.pos >= len(b.b) || b.pos+len(s) > len(b.b) {
		return false
	}
	if !bytes.Equal([]byte(s), b.b[b.pos:b.pos+len(s)]) {
		return false
	}
	b.pos += len(s)
	return true
}

// Uint64 parses the number starting in the buffer at the current position until
// a character other than 0-9 is encountered, or EOL. The number must consist of
// at least a single digit. If successful, Uint64 returns the number and true;
// otherwise zero and false.
func (b *bytestring) Uint64() (num uint64, ok bool) {
	if b.pos >= len(b.b) {
		return 0, false
	}
	ch := b.b[b.pos]
	if ch < '0' || ch > '9' {
		return 0, false
	}
	num = uint64(ch - '0')
	b.pos++
	for {
		if b.pos >= len(b.b) {
			return num, true
		}
		ch := b.b[b.pos]
		if ch < '0' || ch > '9' {
			return num, true
		}
		num = num*10 + uint64(ch-'0')
		b.pos++
	}
}

// NumFields returns the number of fields found in the line, starting from the
// current position. NumFields does not change the current position. Fields are
// made of sequences of characters excluding the space character. Fields are
// separated by one or more spaces.
func (b *bytestring) NumFields() (num int) {
	pos := b.pos
	for {
		for {
			if pos >= len(b.b) {
				return
			}
			if b.b[pos] != ' ' {
				break
			}
			pos++
		}
		num++
		for {
			if pos >= len(b.b) {
				return
			}
			if b.b[pos] == ' ' {
				break
			}
			pos++
		}
	}
}
