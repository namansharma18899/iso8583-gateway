package iso8583

import "testing"

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		pan  string
		want bool
	}{
		{"4532015112830366", true},
		{"4916338506082832", true},
		{"5425233430109903", true},
		{"2221000000000009", true},
		{"4532015112830367", false},
		{"0000000000000000", true},
		{"1234567890123456", false},
		{"123456789012", false},  // too short
		{"12345678901234567890", false}, // too long
		{"45320151128303AB", false}, // non-numeric
	}

	for _, tt := range tests {
		t.Run(tt.pan, func(t *testing.T) {
			if got := ValidateLuhn(tt.pan); got != tt.want {
				t.Errorf("ValidateLuhn(%q) = %v, want %v", tt.pan, got, tt.want)
			}
		})
	}
}
