package multiq

import (
	"testing"
)

var global int

func TestRDRAND(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Logf("%v", rdrand())
	}
}

func BenchmarkRDRAND(b *testing.B) {
	rng := rdrand()
	for i := 0; i < b.N; i++ {
		rng = rdrand()
		global += int(rng)
	}
}

func BenchmarkRNG(b *testing.B) {
	rng := rdrand()
	for i := 0; i < b.N; i++ {
		rng = xorshiftMult64(rng)
		global += int(rng)
	}
}
