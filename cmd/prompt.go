package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

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
