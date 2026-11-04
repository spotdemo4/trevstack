//go:build !dev

package main

import "testing"

func TestBuildAdmitsNoDevelopmentOrigins(t *testing.T) {
	if len(devOrigins) != 0 {
		t.Errorf("devOrigins = %v, want none outside the dev build", devOrigins)
	}
}
