package handlers

import "testing"

func TestValidSLSDashboardProjectOrder(t *testing.T) {
	tests := []struct {
		name string
		ids  []uint
		want bool
	}{
		{name: "valid", ids: []uint{3, 1, 2}, want: true},
		{name: "empty", ids: nil, want: false},
		{name: "zero", ids: []uint{1, 0}, want: false},
		{name: "duplicate", ids: []uint{1, 2, 1}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validSLSDashboardProjectOrder(test.ids); got != test.want {
				t.Fatalf("validSLSDashboardProjectOrder(%v) = %v, want %v", test.ids, got, test.want)
			}
		})
	}
}
