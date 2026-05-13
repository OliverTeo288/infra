package prompt

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

// withStdin replaces the package-level stdinReader for the duration of a test
// and restores it on Cleanup.
func withStdin(t *testing.T, input string) {
	t.Helper()
	original := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader(input))
	t.Cleanup(func() { stdinReader = original })
}

func TestSelection_ReturnsOptionForValidChoice(t *testing.T) {
	withStdin(t, "2\n")
	got, err := Selection([]string{"alpha", "beta", "gamma"}, "thing")
	if err != nil {
		t.Fatalf("Selection: %v", err)
	}
	if got != "beta" {
		t.Errorf("got %q; want beta", got)
	}
}

func TestSelection_EmptyOptionsReturnsErrNoOptions(t *testing.T) {
	_, err := Selection(nil, "anything")
	if !errors.Is(err, ErrNoOptions) {
		t.Errorf("expected ErrNoOptions; got %v", err)
	}
}

func TestSelection_TooManyInvalidAttempts(t *testing.T) {
	withStdin(t, "99\nbanana\n0\n")
	_, err := Selection([]string{"a", "b"}, "")
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Errorf("expected ErrTooManyAttempts; got %v", err)
	}
}

func TestLocalPort_ValidPortAccepted(t *testing.T) {
	withStdin(t, "5432\n")
	port, err := LocalPort()
	if err != nil {
		t.Fatalf("LocalPort: %v", err)
	}
	if port != 5432 {
		t.Errorf("got %d; want 5432", port)
	}
}

func TestLocalPort_RejectsOutOfRange(t *testing.T) {
	withStdin(t, "80\n100\n70000\n")
	_, err := LocalPort()
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Errorf("expected ErrTooManyAttempts; got %v", err)
	}
}

func TestInput_AppliesDefaultOnEmpty(t *testing.T) {
	withStdin(t, "\n")
	got, err := Input("Name", nil, "fallback")
	if err != nil {
		t.Fatalf("Input: %v", err)
	}
	if got != "fallback" {
		t.Errorf("got %q; want fallback", got)
	}
}

func TestInput_RetriesOnValidationFailure(t *testing.T) {
	withStdin(t, "bad\ngood\n")
	got, err := Input("Name", func(s string) error {
		if s == "bad" {
			return errors.New("nope")
		}
		return nil
	}, "")
	if err != nil {
		t.Fatalf("Input: %v", err)
	}
	if got != "good" {
		t.Errorf("got %q; want good", got)
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"\n", false},
		{"maybe\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			withStdin(t, tc.in)
			if got := Confirm("ok?"); got != tc.want {
				t.Errorf("Confirm(%q) = %v; want %v", tc.in, got, tc.want)
			}
		})
	}
}
