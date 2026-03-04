package worker

import "testing"

func TestEvalFormula(t *testing.T) {
	if v := evalFormula("1+2*3", 0); v != 7 {
		t.Fatalf("expected 7 got %v", v)
	}
	if v := evalFormula("10 + t", 5); v != 15 {
		t.Fatalf("expected 15 got %v", v)
	}
}
