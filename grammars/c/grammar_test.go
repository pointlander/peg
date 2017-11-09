package main

import (
	"testing"
)

func parseCBuffer(buffer string) (*C, error) {
	clang := &C{Buffer: buffer}
	clang.Init()
	err := clang.Parse()
	return clang, err
}

func parseC_4t(t *testing.T, src string) *C {
	c, err := parseCBuffer(src)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCParsing_Expressions1(t *testing.T) {
	case1src :=
		`int a() {
		(es);
		1++;
		1+1;
		a+1;
		(a)+1;
		a->x;
		return 0;
}`
	parseC_4t(t, case1src)
}

func TestCParsing_Expressions2(t *testing.T) {
	parseC_4t(t,
		`int a() {
	if (a) { return (a); }

	return (0);
	return a+b;
	return (a+b);
	return (a)+0;
}`)

	parseC_4t(t, `int a() { return (a)+0; }`)
}

func TestCParsing_Expressions3(t *testing.T) {
	parseC_4t(t,
		`int a() {
1+(a);
(a)++;
(es)++;
(es)||a;
(es)->a;
return (a)+(b);
return 0+(a);
}`)
}

func TestCParsing_Expressions4(t *testing.T) {
	parseC_4t(t, `int a(){1+(a);}`)
}
func TestCParsing_Expressions5(t *testing.T) {
	parseC_4t(t, `int a(){return (int)0;}`)
}
func TestCParsing_Expressions6(t *testing.T) {
	parseC_4t(t, `int a(){return (in)0;}`)
}
func TestCParsing_Expressions7(t *testing.T) {
	parseC_4t(t, `int a() 
{ return (0); }`)
}
func TestCParsing_Cast0(t *testing.T) {
	parseC_4t(t, `int a(){(cast)0;}`)
}
func TestCParsing_Cast1(t *testing.T) {
	parseC_4t(t, `int a(){(m*)(rsp);}`)
	parseC_4t(t, `int a(){(struct m*)(rsp);}`)
}

func TestCParsing_Empty(t *testing.T) {
	parseC_4t(t, `/** empty is valid. */  `)
}
