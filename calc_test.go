package main

import "testing"

func TestEvalExpr(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"2 + 3", 5},
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"10 / 2 / 5", 1},
		{"-5 + 10", 5},
		{"3.5 + 1.5", 5},
		{"2 * (3 + (4 - 1))", 12},
	}

	for _, c := range cases {
		got, err := EvalExpr(c.expr)
		if err != nil {
			t.Errorf("EvalExpr(%q) returned error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("EvalExpr(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}

func TestEvalExprErrors(t *testing.T) {
	cases := []string{
		"",
		"2 +",
		"(2 + 3",
		"2 / 0",
		"2 $ 3",
	}

	for _, expr := range cases {
		if _, err := EvalExpr(expr); err == nil {
			t.Errorf("EvalExpr(%q) expected an error, got none", expr)
		}
	}
}
