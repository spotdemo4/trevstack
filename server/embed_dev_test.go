//go:build !prod

package main

import (
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

func TestDevBuildAdmitsTheWebDevelopmentServer(t *testing.T) {
	admit := trevrpc.SameOriginAdmission(devOrigins...)
	for _, origin := range []string{"http://localhost:3000", "http://127.0.0.1:3000"} {
		if !admit(trevrpc.AdmissionRequest{Origin: origin, Authority: "localhost:8080", Secure: true}) {
			t.Errorf("dev build does not admit %s", origin)
		}
	}
}
