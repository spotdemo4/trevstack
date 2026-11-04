//go:build dev

package main

// The web app's development server serves it instead of this build, so its
// pages may call the API.
func init() {
	devOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
}
