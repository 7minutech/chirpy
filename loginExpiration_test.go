package main

import "testing"

func TestSetExperation(t *testing.T) {
	cases := []struct {
		input    int
		expected int
	}{
		{
			input:    0,
			expected: 3600,
		},
		{
			input:    3601,
			expected: 3600,
		},
		{
			input:    240,
			expected: 240,
		},
	}

	for _, c := range cases {
		actual := setExperation(c.input)
		if actual != c.expected {
			t.Errorf("setExperation(%d) == %d, expected: %d", c.input, actual, c.expected)
		}
	}
}
