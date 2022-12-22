package execext

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func CommandContextStream(ctx context.Context, name string, args ...string) (string, error) {
	var (
		output []string
		errors []string
	)
	cmd := exec.CommandContext(ctx, name, args...)
	{
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", err
		}
		go func(scanner *bufio.Scanner) {
			for scanner.Scan() {
				output = append(output, scanner.Text())
			}
		}(bufio.NewScanner(stdout))
	}
	{
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return "", err
		}
		go func(scanner *bufio.Scanner) {
			for scanner.Scan() {
				errors = append(errors, scanner.Text())
			}
		}(bufio.NewScanner(stderr))
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf(strings.Join(errors, "\n"))
	}
	return strings.Join(output, "\n"), nil
}
