package cmd

import "testing"

func TestNormalizeImageSizeMatchesDocumentedRange(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "below minimum", in: 15, want: 20},
		{name: "minimum", in: 20, want: 20},
		{name: "interior", in: 35, want: 35},
		{name: "above old maximum", in: 40, want: 40},
		{name: "maximum", in: 50, want: 50},
		{name: "above maximum", in: 51, want: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeImageSize(tt.in); got != tt.want {
				t.Errorf("normalizeImageSize(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
