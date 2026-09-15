package handlers

import "testing"

func TestDerivedViewFilterContract(t *testing.T) {
	for _, facet := range []string{"", "business", "uiux", "frontend", "backend"} {
		if !validDerivedFacet(facet) {
			t.Fatalf("expected %q to be valid", facet)
		}
	}
	if validDerivedFacet("unknown") {
		t.Fatal("unknown facet must be rejected")
	}
	marks, args := placeholders([]string{"wi_1", "wi_2"}, 3)
	if marks != "$3,$4" || len(args) != 2 {
		t.Fatalf("unexpected placeholders: %q %#v", marks, args)
	}
}
