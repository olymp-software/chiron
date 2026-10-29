package runtime

import (
	"slices"
	"sync"
	"testing"
	"time"
)

func TestDecoderSplitCharacter(t *testing.T) {
	var d Decoder
	euro := []byte("€") // three bytes
	got := d.Decode(append([]byte("a"), euro[:1]...))
	got += "|" + d.Decode(euro[1:2])
	got += "|" + d.Decode(append(euro[2:], 'b'))
	if want := "a||€b"; got != want {
		t.Errorf("decoded = %q, want %q", got, want)
	}
}

func TestDecoderInvalidBytes(t *testing.T) {
	var d Decoder
	if got, want := d.Decode([]byte("a\xffb\xc0c")), "a�b�c"; got != want {
		t.Errorf("decoded = %q, want %q", got, want)
	}
}

func TestDecoderFlushIncomplete(t *testing.T) {
	var d Decoder
	got := d.Decode([]byte("a\xe2\x82"))
	got += "|" + d.Flush()
	if want := "a|�"; got != want {
		t.Errorf("decoded = %q, want %q", got, want)
	}
}

// flushes collects what a Batcher passes on.
type flushes struct {
	mu   sync.Mutex
	list []string
}

func (f *flushes) add(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.list = append(f.list, s)
}

func (f *flushes) get() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.list)
}

func TestBatcherFlushesOnTick(t *testing.T) {
	tick := make(chan time.Time)
	var f flushes
	b := newBatcher(tick, f.add)
	_, _ = b.Write([]byte("a"))
	_, _ = b.Write([]byte("b"))
	if got := f.get(); len(got) != 0 {
		t.Errorf("flushed before a tick: %q", got)
	}
	tick <- time.Time{}
	eventually(t, "first batch", func() bool { return len(f.get()) == 1 })
	tick <- time.Time{} // nothing new: no flush
	_, _ = b.Write([]byte("c\xe2\x82"))
	tick <- time.Time{}
	eventually(t, "second batch", func() bool { return len(f.get()) == 2 })
	_, _ = b.Write([]byte("\xac"))
	b.Close()
	if got, want := f.get(), []string{"ab", "c", "€"}; !slices.Equal(got, want) {
		t.Errorf("batches = %q, want %q", got, want)
	}
}

func TestBatcherCloseFlushesRest(t *testing.T) {
	var f flushes
	b := newBatcher(make(chan time.Time), f.add)
	_, _ = b.Write([]byte("x\xe2"))
	b.Close()
	if got, want := f.get(), []string{"x�"}; !slices.Equal(got, want) {
		t.Errorf("batches = %q, want %q", got, want)
	}
}
