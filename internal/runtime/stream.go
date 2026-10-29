package runtime

import (
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// BatchInterval is how often a Batcher passes collected output on.
const BatchInterval = 75 * time.Millisecond

// Decoder turns a stream of bytes into valid UTF-8 text. A multi-byte
// character split across writes is held back until it is complete; invalid
// bytes become U+FFFD.
type Decoder struct {
	pending []byte
}

// Decode returns the text of p and the bytes held back from earlier calls.
func (d *Decoder) Decode(p []byte) string {
	buf := append(d.pending, p...)
	cut := incompleteTail(buf)
	d.pending = append([]byte(nil), buf[cut:]...)
	return strings.ToValidUTF8(string(buf[:cut]), string(utf8.RuneError))
}

// Flush returns the bytes held back, which can no longer be completed.
func (d *Decoder) Flush() string {
	s := strings.ToValidUTF8(string(d.pending), string(utf8.RuneError))
	d.pending = nil
	return s
}

// incompleteTail returns where an incomplete UTF-8 sequence at the end of b
// starts, or len(b) if there is none.
func incompleteTail(b []byte) int {
	for i := len(b) - 1; i >= 0 && i >= len(b)-utf8.UTFMax+1; i-- {
		if utf8.RuneStart(b[i]) {
			if !utf8.FullRune(b[i:]) {
				return i
			}
			break
		}
	}
	return len(b)
}

// Batcher is a writer that decodes output and passes it to flush at most
// once per tick, so endless output costs a bounded number of UI events.
type Batcher struct {
	flush    func(string)
	stopTick func()
	stop     chan struct{}
	done     chan struct{}

	mu  sync.Mutex
	dec Decoder
	buf strings.Builder
}

// NewBatcher returns a Batcher that flushes every BatchInterval.
func NewBatcher(flush func(string)) *Batcher {
	t := time.NewTicker(BatchInterval)
	b := newBatcher(t.C, flush)
	b.stopTick = t.Stop
	return b
}

func newBatcher(tick <-chan time.Time, flush func(string)) *Batcher {
	b := &Batcher{flush: flush, stopTick: func() {}, stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(b.done)
		for {
			select {
			case <-tick:
				b.send()
			case <-b.stop:
				return
			}
		}
	}()
	return b
}

func (b *Batcher) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.WriteString(b.dec.Decode(p))
	return len(p), nil
}

// Close passes on what is left and stops the Batcher; flush is not called
// after Close returns.
func (b *Batcher) Close() {
	b.stopTick()
	close(b.stop)
	<-b.done
	b.mu.Lock()
	b.buf.WriteString(b.dec.Flush())
	b.mu.Unlock()
	b.send()
}

func (b *Batcher) send() {
	b.mu.Lock()
	s := b.buf.String()
	b.buf.Reset()
	b.mu.Unlock()
	if s != "" {
		b.flush(s)
	}
}
