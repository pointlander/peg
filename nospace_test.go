package main

import "testing"

func TestNoSpacePackage(t *testing.T) {
	buffer := `packagenospace
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err == nil {
		t.Error("packagenospace was parsed without error")
	}
}

func TestNoSpaceType(t *testing.T) {
	buffer := `
package p
typenospace Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err == nil {
		t.Error("typenospace was parsed without error")
	}
}
