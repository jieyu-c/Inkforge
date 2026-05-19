package phone

import "testing"

func TestMaskDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{"+8615012345678", "+86150****5678"},
		{"8615012345678", "+86150****5678"},
		{"+86 150 1234 5678", "+86150****5678"},
		{"15012345678", "+86150****5678"},
		{"+8613876543210", "+86138****3210"},
		{"xyz", "***"},
		{"", "***"},
		{"12", "**"},
		{"12345678901234", "**********1234"}, // fallback: tail 4
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := MaskDisplay(tt.in); got != tt.want {
				t.Errorf("MaskDisplay(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
