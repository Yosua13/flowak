package middleware

import "testing"

func TestProjectAuthorizationMatrix(t *testing.T) {
	cases := []struct {
		role       string
		capability ProjectCapability
		want       bool
	}{
		{"owner", CapabilityManage, true},
		{"owner", CapabilityEditGraph, true},
		{"editor", CapabilityEditGraph, true},
		{"editor", CapabilityManage, false},
		{"commenter", CapabilityWorkItem, true},
		{"commenter", CapabilityEditGraph, false},
		{"viewer", CapabilityView, true},
		{"viewer", CapabilityComment, false},
		{"viewer", CapabilityEditGraph, false},
		{"viewer", CapabilityManage, false},
	}
	for _, tc := range cases {
		if got := can(tc.role, tc.capability); got != tc.want {
			t.Errorf("can(%q, %q) = %v, want %v", tc.role, tc.capability, got, tc.want)
		}
	}
}

func TestUnknownProjectRoleHasNoCapabilities(t *testing.T) {
	for _, capability := range []ProjectCapability{CapabilityView, CapabilityComment, CapabilityWorkItem, CapabilityEditGraph, CapabilityManage} {
		if can("pm", capability) || can("", capability) {
			t.Fatalf("functional or empty roles must not authorize %q", capability)
		}
	}
}
