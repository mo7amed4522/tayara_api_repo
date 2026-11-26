package utils

import "testing"

func TestHashChainDeterministicAndVerifiable(t *testing.T) {
	h1 := HashChainStep(HashChainZero, "first")
	h2 := HashChainStep(h1, "second")
	if h1 == "" || h1 == h2 {
		t.Fatal("chain links must differ")
	}
	if !VerifyChain([]string{"first", "second"}, h2) {
		t.Fatal("valid chain failed verification")
	}
	if VerifyChain([]string{"first", "tampered"}, h2) {
		t.Fatal("tampered chain verified")
	}
	if again := HashChainStep(HashChainZero, "first"); again != h1 {
		t.Fatal("hash must be deterministic")
	}
}
