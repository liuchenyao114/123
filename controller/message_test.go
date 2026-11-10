package controller

import "testing"

func TestSplitChannelNames(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"single", "bark", []string{"bark"}},
		{"two", "bark|cmcc", []string{"bark", "cmcc"}},
		{"with_spaces", " bark | cmcc | corp ", []string{"bark", "cmcc", "corp"}},
		{"empty_segments", "bark||cmcc|", []string{"bark", "cmcc"}},
		{"duplicates", "bark|bark|cmcc", []string{"bark", "cmcc"}},
		{"only_separators", "||", []string{}},
		{"empty", "", []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitChannelNames(tc.raw)
			if len(got) != len(tc.want) {
				t.Fatalf("splitChannelNames(%q) = %v, want %v", tc.raw, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("splitChannelNames(%q)[%d] = %q, want %q", tc.raw, i, got[i], tc.want[i])
				}
			}
		})
	}
}
