package handlers

import "testing"

func TestGraphCompatibilityWriteDefaultsOffAndSupportsRollbackFlag(t *testing.T) {
	t.Setenv("GRAPH_COMPATIBILITY_WRITE_ENABLED", "")
	if graphCompatibilityWriteEnabled() {
		t.Fatal("compatibility writes must default off")
	}
	t.Setenv("GRAPH_COMPATIBILITY_WRITE_ENABLED", "true")
	if !graphCompatibilityWriteEnabled() {
		t.Fatal("rollback flag must re-enable compatibility writes")
	}
}
