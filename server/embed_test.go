//go:build prod

package main

import "testing"

func TestBuildAdmitsNoDevelopmentOrigins(t *testing.T) {
	if len(devOrigins) != 0 {
		t.Errorf("devOrigins = %v, want none in the prod build", devOrigins)
	}
}
