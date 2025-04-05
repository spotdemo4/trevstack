package utils

import (
	"bufio"
	"os/exec"
)

type Stdout string
type Stderr string
type ExitCode int

func Run(cmd *exec.Cmd) (chan interface{}, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	c := make(chan interface{}, 10)

	go func() {
		scan := bufio.NewScanner(stdout)
		for scan.Scan() {
			c <- Stdout(scan.Text())
		}
	}()
	go func() {
		scan := bufio.NewScanner(stderr)
		for scan.Scan() {
			c <- Stderr(scan.Text())
		}
	}()

	go func() {
		defer close(c)

		if err := cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				c <- ExitCode(exitError.ExitCode())
			} else {
				c <- ExitCode(1)
			}
		} else {
			c <- ExitCode(0)
		}
	}()

	return c, nil
}
