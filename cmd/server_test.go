package cmd

import "testing"

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: Run() executes the root command / server")
	}
	Run()
}
