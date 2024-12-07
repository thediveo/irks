/*

## AllCounters

go test -bench=AllCounters -run=^$ -cpu=1,4,8,16 -benchmem -benchtime=15s

goos: linux
goarch: amd64
pkg: github.com/thediveo/irks
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkAllCounters               65942            278857 ns/op            4736 B/op          6 allocs/op
BenchmarkAllCounters-4             65719            271429 ns/op            4736 B/op          6 allocs/op
BenchmarkAllCounters-8             66358            270766 ns/op            4736 B/op          6 allocs/op
BenchmarkAllCounters-16            66622            269355 ns/op            4736 B/op          6 allocs/op

...not unexpectedly, the number of CPUs doesn't really change a thing, as there
is no concurrency going on when parsing /proc/interrupts. Only 6 allocations per
iteration looks like we reached our goal of keeping allocations very few.

## AllDetails

go test -bench=AllDetails -run=^$ -cpu=1,4,8,16 -benchmem -benchtime=15s

Using io.ReadFile to fetch the IRQ meta data and CPU effective affinities list,
we get the following figures:

goos: linux
goarch: amd64
pkg: github.com/thediveo/irks
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkAllDetails                13435           1322222 ns/op          633459 B/op       1296 allocs/op
BenchmarkAllDetails-4              13232           1349680 ns/op          634238 B/op       1296 allocs/op
BenchmarkAllDetails-8              13305           1298408 ns/op          634175 B/op       1296 allocs/op
BenchmarkAllDetails-16             13096           1364539 ns/op          634427 B/op       1296 allocs/op

Now, let's replace io.ReadFile with our own readFile that recycles the read
buffer and simply overrides its previous contents – which is perfectly fine as
we're making sure we don't keep any references to the old contents.

goos: linux
goarch: amd64
pkg: github.com/thediveo/irks
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkAllDetails                17356           1050542 ns/op           42273 B/op        989 allocs/op
BenchmarkAllDetails-4              16946           1050304 ns/op           42373 B/op        989 allocs/op
BenchmarkAllDetails-8              17127           1045725 ns/op           42383 B/op        989 allocs/op
BenchmarkAllDetails-16             17091           1061204 ns/op           42408 B/op        989 allocs/op

...that's almost 24% less allocations, or almost a quarter less. We successfully
managed to massacre the amount of allocated memory needed to iterate the
details, this is down to(!) only 7% of memory allocated compared to the
io.ReadFile implementation. That's "down to", not "by". We've also got 20%
faster.

*/

package irks_test

import (
	"testing"

	"github.com/thediveo/irks"
)

func BenchmarkAllCounters(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// Note https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go
		// on compiler optimizations: as of go 1.23.4 we seem to be on a safe
		// side here so far as things get pushed to us and it doesn't matter if
		// we copy or not.
		for range irks.AllCounters() {
		}
	}
}

func BenchmarkAllDetails(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range irks.AllIRQDetails() {
		}
	}
}
