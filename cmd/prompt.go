package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// promptContinue mirrors gcloud-python's
// `console_io.PromptContinue(message, cancel_on_no=True, cancel_string=...)`
// with the default answer of Yes. It writes the message and a
// "Do you want to continue (Y/n)?" prompt to stderr, and returns an error
// carrying cancelString when the user answers no. With --quiet, or when stdin
// is not a terminal, it returns nil without prompting.
func promptContinue(message, cancelString string) error {
	if flagQuiet || !IsInteractive() {
		return nil
	}
	return promptContinueIO(os.Stdin, os.Stderr, message, cancelString)
}

// promptContinueIO is the io-taking form of promptContinue, used by tests.
func promptContinueIO(in io.Reader, out io.Writer, message, cancelString string) error {
	if message != "" {
		fmt.Fprintf(out, "%s\n\n", message)
	}
	fmt.Fprint(out, "Do you want to continue (Y/n)?  ")
	defer fmt.Fprintln(out)

	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		switch {
		case answer == "y" || answer == "yes":
			return nil
		case answer == "n" || answer == "no":
			if cancelString == "" {
				cancelString = "Aborted by user."
			}
			return errors.New(cancelString)
		case answer == "" || err != nil:
			// Empty input (or EOF) takes the default, Yes.
			return nil
		}
		fmt.Fprint(out, "Please enter 'y' or 'n':  ")
	}
}
