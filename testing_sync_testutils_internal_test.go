package core

// The channel setups carry int values, numbered from 1 in send order, so a
// row can declare what an assertion hands back.

// ChanOpenForTest returns a channel nothing will close or send to.
func ChanOpenForTest() <-chan int {
	return make(chan int)
}

// ChanClosedForTest returns an already closed channel.
func ChanClosedForTest() <-chan int {
	ch := make(chan int)
	close(ch)
	return ch
}

// ChanNilForTest returns a nil channel, which never closes.
func ChanNilForTest() <-chan int {
	return nil
}

// ChanWithSendsForTest returns a setup that fills a buffered channel
// with n values before handing it over.
func ChanWithSendsForTest(n int) func() <-chan int {
	return func() <-chan int {
		ch := make(chan int, n)
		for i := range n {
			ch <- i + 1
		}
		return ch
	}
}

// ChanFirstTakenForTest returns a channel filled with three values and
// the first of them already taken, so the next one read is 2.
func ChanFirstTakenForTest() <-chan int {
	ch := ChanWithSendsForTest(3)()
	<-ch
	return ch
}

// ChanClosedWithSendsForTest returns a setup that fills a buffered
// channel with n values and closes it, so the close sits behind the
// values.
func ChanClosedWithSendsForTest(n int) func() <-chan int {
	return func() <-chan int {
		ch := make(chan int, n)
		for i := range n {
			ch <- i + 1
		}
		close(ch)
		return ch
	}
}
