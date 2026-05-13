// Package prompt provides interactive stdin prompts used by the CLI. All
// prompts share a single bufio.Reader to avoid losing buffered input when
// callers switch between selection / confirmation / free-text prompts.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	maxAttempts = 3
	minPort     = 1024
	maxPort     = 65535
)

// ErrNoOptions is returned when a selection prompt is invoked with an empty
// option list. Callers can use errors.Is to distinguish from validation errors.
var ErrNoOptions = errors.New("no options available to select from")

// ErrTooManyAttempts is returned after maxAttempts invalid inputs.
var ErrTooManyAttempts = errors.New("too many invalid attempts")

// stdinReader is a process-wide reader for stdin. Reusing the same reader
// avoids losing buffered bytes when callers create new bufio.Readers
// against os.Stdin between prompts.
var stdinReader = bufio.NewReader(os.Stdin)

func readLine() (string, error) {
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// Selection prompts the user to pick one of options. taskName is shown in
// the prompt header; pass "" for a generic header.
func Selection(options []string, taskName string) (string, error) {
	if len(options) == 0 {
		return "", ErrNoOptions
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if taskName != "" {
			fmt.Printf("Please select %s from the following options:\n", taskName)
		} else {
			fmt.Println("Please select from the following options:")
		}
		for i, opt := range options {
			fmt.Printf("[%d] %s\n", i+1, opt)
		}

		fmt.Print("Enter the number of your choice: ")
		in, err := readLine()
		if err != nil {
			return "", err
		}
		if i, err := strconv.Atoi(in); err == nil && i > 0 && i <= len(options) {
			return options[i-1], nil
		}

		if remaining := maxAttempts - attempt; remaining > 0 {
			fmt.Printf("Invalid choice. You have %d attempt(s) remaining.\n", remaining)
		}
	}
	return "", ErrTooManyAttempts
}

// LocalPort prompts for a TCP port in [minPort, maxPort].
func LocalPort() (int, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("Enter a local port number for port forwarding (%d-%d): ", minPort, maxPort)
		in, err := readLine()
		if err != nil {
			return 0, err
		}
		if port, err := strconv.Atoi(in); err == nil && port >= minPort && port <= maxPort {
			return port, nil
		}
		if remaining := maxAttempts - attempt; remaining > 0 {
			fmt.Printf("Invalid port number. You have %d attempt(s) remaining.\n", remaining)
		}
	}
	return 0, ErrTooManyAttempts
}

// Input prompts the user for free-form text, optionally validating and/or
// applying a default when the user submits an empty line.
func Input(label string, validate func(string) error, defaultValue string) (string, error) {
	header := label + ": "
	if defaultValue != "" {
		header = fmt.Sprintf("%s [%s]: ", label, defaultValue)
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Print(header)
		in, err := readLine()
		if err != nil {
			return "", err
		}
		if in == "" && defaultValue != "" {
			in = defaultValue
		}
		if validate != nil {
			if err := validate(in); err != nil {
				fmt.Printf("Invalid input: %s\n", err)
				continue
			}
		}
		return in, nil
	}
	return "", ErrTooManyAttempts
}

// Confirm displays a yes/no prompt; returns true only on a "y" / "yes" response.
func Confirm(message string) bool {
	fmt.Print(message + " ")
	in, err := readLine()
	if err != nil {
		fmt.Println("Error reading input:", err)
		return false
	}
	response := strings.ToLower(in)
	return response == "y" || response == "yes"
}
