package controller

import "testing"

func TestAuthMessage(t *testing.T) {
	ptr := func(s string) *string { return &s }
	cases := []struct {
		name     string
		msgToken string
		usrToken string
		chToken  *string
		want     bool
	}{
		{"no token configured at all", "", "", nil, true},
		{"global token matches", "G", "G", nil, true},
		{"global token set but request has none", "", "G", nil, false},
		{"global token set, wrong token", "X", "G", nil, false},
		{"channel token matches", "A", "", ptr("A"), true},
		{"channel token set but request has none", "", "", ptr("A"), false},
		{"channel token wrong", "B", "", ptr("A"), false},
		{"global token matches with channel token", "G", "G", ptr("A"), true},
		{"channel token matches with global token set", "A", "G", ptr("A"), true},
		{"global token set, channel token empty", "", "G", ptr(""), false},
		{"no global token, channel token empty", "", "", ptr(""), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := authMessage(tc.msgToken, tc.usrToken, tc.chToken)
			if got != tc.want {
				t.Fatalf("authMessage(%q, %q, chToken) = %v, want %v",
					tc.msgToken, tc.usrToken, got, tc.want)
			}
		})
	}
}

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
