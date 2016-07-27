package multiq

import (
	"testing"
)

var global int

func TestRDTSC(t *testing.T) {

	ticks := rdtsc()

	for i := 0; i < 10; i++ {
		global += i
	}

	t.Logf("10 ticks=%d", rdtsc()-ticks)

	ticks = rdtsc()

	for i := 0; i < 100; i++ {
		global += i
	}

	t.Logf("100 ticks=%d", rdtsc()-ticks)
}
