package stringutil

import "testing"

func TestIsEmpty(t *testing.T) {
	cases := map[string]struct {
		in   string
		want bool
	}{
		"empty":        {"", true},
		"single space": {" ", false},
		"word":         {"go", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsEmpty(tc.in); got != tc.want {
				t.Errorf("IsEmpty(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	cases := map[string]string{
		"":        "",
		"a":       "a",
		"ab":      "ba",
		"racecar": "racecar",
		"hello":   "olleh",
	}
	for in, want := range cases {
		if got := Reverse(in); got != want {
			t.Errorf("Reverse(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCountVowels(t *testing.T) {
	cases := map[string]int{
		"":                0,
		"xyz":             0,
		"aeiou":           5,
		"AEIOU":           5,
		"Hello, World!":   3,
		"The Quick Brown": 4,
	}
	for in, want := range cases {
		if got := CountVowels(in); got != want {
			t.Errorf("CountVowels(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestIsPalindrome(t *testing.T) {
	cases := map[string]bool{
		"":        true,
		"a":       true,
		"racecar": true,
		"abba":    true,
		"hello":   false,
		"ab":      false,
	}
	for in, want := range cases {
		if got := IsPalindrome(in); got != want {
			t.Errorf("IsPalindrome(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello..."},
		{"", 3, ""},
	}
	for _, tc := range cases {
		if got := Truncate(tc.in, tc.max); got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}
