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
	"sync"

	"github.com/thediveo/faf"
)

// IRQDetails provides the list of actions and the currently set CPU affinities
// for a specific IRQ, as indicated by Num.
type IRQDetails struct {
	Num        uint          // IRQ number
	Actions    string        // list of IRQ actions
	Affinities CPUAffinities // effective CPU(s) affinities
}

// CPUAffinities is a list of CPU [from...to] ranges. CPU numbers are starting
// from zero.
type CPUAffinities [][2]uint

// AllIRQDetails returns an iterator looping over the details of all
// (non-architecture-specific) IRQs in the system, giving their details as to
// actions and CPU affinities.
//
// AllIRQDetails uses a streamlined implementation that runs at approx 1.8× the
// execution speed compared to a “traditional” Go implementation approach using
// [os.File.ReadDir], [strconv.ParseUint] and [os.ReadFile]. For the same system
// with 47 hardware IRQs, we only need at around 10% of memory on the heap, and
// only 1/3 of allocations, compared to using stock stdlib functions.
func AllIRQDetails() iter.Seq[IRQDetails] {
	return allIRQDetails("")
}

const (
	syskernelirqPath = "/sys/kernel/irq/"
	procirqPath      = "/proc/irq/"

	actionsNode           = "/actions"
	effectiveAffinityNode = "/effective_affinity_list"
)

const size = 16

// allIRQDetails loops over the details of the IRQs available in this
// host/system.
//
// The iterator returned aggressively optimizes speed by reading the many small
// pseudo files with the sought-after details in concurrent worker go routines.
// This allows the Linux kernel to render the contents of these pseudo files on
// multiple CPUs simultaneously, avoiding a single CPU thread bottleneck.
func allIRQDetails(root string) iter.Seq[IRQDetails] {
	return func(yield func(IRQDetails) bool) {
		// Closing done signals to workers to prematurely terminate, even if
		// there would be still more details to be looked up; this signal is
		// emitted when the yield function returns false.
		done := make(chan struct{})
		// The job queue with the names (numbers) of IRQs to look their details
		// up. Closing this job queue signals the workers after draining this
		// channel that all work is done and they need to terminate.
		namech := make(chan string, size)
		// The multi-producer-single-consumer queue gathering the IRQ details to
		// be pushed to the yield function. As this is a multi-producer channel,
		// it must be closed only after all workers have terminated, or
		// otherwise panic ensues.
		detailch := make(chan IRQDetails, size)
		// We need to signal the loop that pulls off IRQ details and feeds them
		// into the yield function that all details have been consumed: for this
		// we close the detail channel, but we can do so only after all details
		// have been posted into the details channel by our workers. So we track
		// the number of workers that haven't already terminated with a wait
		// group. Only after this worker wait group is done, we can close the
		// detail channel without any danger of a late worker panicking.
		var wg sync.WaitGroup

		// The worker function takes IRQ numbers/names off the name channel and
		// then fetches the detail information about that particular IRQ. It
		// then posts its findings into the detail channel.
		//
		// A worker quits when the name channel after all remaining names have
		// been pulled off the closed name channel.
		//
		// In addition, a worker also quits when the done channel gets closed;
		// this signals that the yield function won't take it anymore.
		readDetails := func() {
			defer wg.Done() // don't count me in anymore after I've gone!
			var name string
			var ok bool
			for {
				select {
				case <-done: // yield function told us to stop it.
					return
				case name, ok = <-namech:
					if !ok {
						return
					}
				}
				// Using bytes.Buffer instead of assembling path strings piecewise
				// doesn't buy us anything above the noise floor, even with
				// preallocating the buffer's capacity once and then truncating back to
				// the root. But reusing the buffer to read the pseudo files boosts us...
				var contents []byte
				var details IRQDetails

				irqnum, ok := faf.ParseUint([]byte(name))
				if !ok {
					continue
				}
				details.Num = uint(irqnum)

				contents, ok = faf.ReadFile(
					root+syskernelirqPath+name+actionsNode, contents)
				if !ok || len(contents) < 1 || contents[len(contents)-1] != '\n' {
					continue
				}
				details.Actions = string(contents[:len(contents)-1]) // escapes

				contents, ok = faf.ReadFile(
					root+procirqPath+name+effectiveAffinityNode, contents)
				if !ok || len(contents) < 1 || contents[len(contents)-1] != '\n' {
					continue
				}
				afflist := cpuList(contents[:len(contents)-1])
				if len(afflist) == 0 {
					continue
				}
				details.Affinities = afflist
				detailch <- details
			}
		}
		// Kick off a number of workers that then wait on the name channel for
		// IRQ "names" (actually, numbers in text format) to become available.
		wg.Add(size)
		for i := 0; i < size; i++ {
			go readDetails()
		}
		// Next, kick off another go routine that reads the available IRQs
		// (numbers) and feeds them into the name channel where the workers wait
		// to drain it and do their work. The name channel is closed after the
		// last IRQ number has been put into it, in order to make the workers
		// terminate when all is said and done.
		go func() {
			for irqEntry := range faf.ReadDir(root + syskernelirqPath) {
				if !irqEntry.IsDir() {
					continue
				}
				namech <- string(irqEntry.Name)
			}
			close(namech)
		}()
		// We wait for all workers to have terminated and then close the detail
		// channel to signal to the iterator to finally terminate. Normally,
		// workers terminate after the name channel (which acts as a job queue)
		// was closed and all pending names drained. But workers can also
		// terminate prematurely in case the yield function returns false; in
		// this situation, workers will instead terminate based on the signal
		// from the closed done channel. Closing the job channel in this case
		// doesn't hurt, as we're the only place here where we close the names
		// (jobs) channel.
		go func() {
			wg.Wait()
			close(detailch)
		}()
		// Now pick up the IRQ details as they are produced by the workers and
		// feed them sequentially to the yield function. If the yield function
		// indicates a premature end, we signal the workers to wind down by
		// closing the done channel. Details that are still in the buffered
		// details channel will eventually be garbage collected.
		for {
			details := <-detailch
			if details.Actions == "" {
				break
			}
			if !yield(details) {
				close(done)
				return
			}
		}
	}
}

// cpuList returns the CPUAffinities list from the given byte slice. Passing a
// byte slice instead of a string avoids any potentially costly conversions from
// mutable byte slices to immutable strings out of the game without really
// complicating things in this case.
func cpuList(b []byte) CPUAffinities {
	bstr := faf.NewBytestring(b)
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
