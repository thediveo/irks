/*

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

go test -bench=AllDetails -run=^$ -cpu=1,4,8,16 -benchmem -benchtime=15s

goos: linux
goarch: amd64
pkg: github.com/thediveo/irks
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkAllDetails                13435           1322222 ns/op          633459 B/op       1296 allocs/op
BenchmarkAllDetails-4              13232           1349680 ns/op          634238 B/op       1296 allocs/op
BenchmarkAllDetails-8              13305           1298408 ns/op          634175 B/op       1296 allocs/op
BenchmarkAllDetails-16             13096           1364539 ns/op          634427 B/op       1296 allocs/op

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
