package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var defaultRsyncOptions = []string{
	"--rsync-path=/usr/local/sbin/rsyncplan-exechook",
	"--timeout=1200",
	"--exclude=/swapfile",
	"-iSalx",
}

type rsyncOptions []string

func (o *rsyncOptions) String() string {
	return strings.Join(*o, " ")
}

func (o *rsyncOptions) Set(value string) error {
	if value == "" {
		return errors.New("rsync argument cannot be empty")
	}
	*o = append(*o, value)
	return nil
}

func loadRsyncOptions(path string) ([]string, error) {
	explicit := path != ""
	if !explicit {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("cannot find default rsync options file: %v; using built-in options", err)
			return append([]string(nil), defaultRsyncOptions...), nil
		}
		path = filepath.Join(home, ".rsyncplan", "config.txt")
	}

	file, err := os.Open(path)
	if err != nil {
		if !explicit && errors.Is(err, os.ErrNotExist) {
			return append([]string(nil), defaultRsyncOptions...), nil
		}
		return nil, fmt.Errorf("open rsync options file %q: %w", path, err)
	}
	defer file.Close()

	var options []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			options = append(options, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read rsync options file %q: %w", path, err)
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("rsync options file %q contains no arguments", path)
	}
	return options, nil
}
