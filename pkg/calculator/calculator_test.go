package calculator

import (
	"testing"
)

type calcTest struct {
	Input string
	Result string
}

func TestParseEval(t *testing.T) {
	tests := []calcTest {
		calcTest { "2+3", "5" },
	}
	for _, test := range tests {
		calc := &Calculator{Buffer: test.Input}
		calc.Init()
		calc.Expression.Init(test.Input)
		if err := calc.Parse(); err != nil {
			t.Errorf("Parse failed: %v", err)
		}
		result := calc.Evaluate().String()
		if result != test.Result {
			t.Errorf("Evaluate('%s') failed. Want: [%s], has: [%s]", test.Result, result)
		}
	}
}