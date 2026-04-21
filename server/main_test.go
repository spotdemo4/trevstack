package main

import "testing"

func TestHello(t *testing.T) {
	expected := "Hello, World!"
	result := hello()
	if result != expected {
		t.Errorf("hello() = %q; want %q", result, expected)
	}
}
