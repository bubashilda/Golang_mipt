//go:build !solution

package waitgroup

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup struct {
	notification chan struct{}
	counter      chan int
	waiters      chan int
}

// New creates WaitGroup.
func New() *WaitGroup {
	wg := WaitGroup{
		notification: make(chan struct{}),
		counter:      make(chan int, 1),
		waiters:      make(chan int, 1),
	}
	wg.counter <- 0
	wg.waiters <- 0
	return &wg
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls have returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	var counter = <-wg.counter

	counter += delta
	if counter < 0 {
		panic("negative WaitGroup counter")
	} else if counter == 0 {
		waiters := <-wg.waiters
		for i := 0; i < waiters; i++ {
			wg.notification <- struct{}{}
		}
		wg.waiters <- 0
	}

	wg.counter <- counter
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	waiters := <-wg.waiters
	wg.waiters <- waiters + 1
	<-wg.notification
}
