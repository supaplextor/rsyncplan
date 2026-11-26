package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadRsyncOptions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	options, err := loadRsyncOptions("")
	if err != nil || !reflect.DeepEqual(options, defaultRsyncOptions) {
		t.Fatalf("default options = %q, error = %v", options, err)
	}
	options[0] = "changed"
	if defaultRsyncOptions[0] == "changed" {
		t.Fatal("returned options modified the defaults")
	}

	dir := filepath.Join(home, ".rsyncplan")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.txt")
	if err := os.WriteFile(path, []byte(" # comment\n\n -iSalx \n--exclude=/path with spaces\n"), 0600); err != nil {
		t.Fatal(err)
	}
	options, err = loadRsyncOptions("")
	if err != nil || !reflect.DeepEqual(options, []string{"-iSalx", "--exclude=/path with spaces"}) {
		t.Fatalf("configured options = %q, error = %v", options, err)
	}

	if _, err := loadRsyncOptions(filepath.Join(home, "missing")); err == nil {
		t.Fatal("missing explicit file did not fail")
	}
	if err := os.WriteFile(path, []byte("# no arguments\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRsyncOptions(""); err == nil {
		t.Fatal("empty config did not fail")
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 70*1024)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRsyncOptions(""); err == nil {
		t.Fatal("unreadable config line did not fail")
	}
}

func TestRsyncOptionsIntegration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("PATH", dir)
	t.Setenv("RSYNC_ARGS", filepath.Join(dir, "rsync-args"))
	for name, script := range map[string]string{
		"ssh":   "#!/bin/sh\nprintf '%s\\n' /backups/example/rootfs/2025-01-01_000000.000000000/\n",
		"rsync": "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$RSYNC_ARGS\"\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}

	run := func(args ...string) []byte {
		t.Helper()
		cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestMainHelper$", "--"}, args...)...)
		cmd.Env = append(os.Environ(), "RSYNCPLAN_TEST_HELPER=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("rsyncplan: %v\n%s", err, output)
		}
		data, err := os.ReadFile(filepath.Join(dir, "rsync-args"))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	defaultArgs := strings.Split(strings.TrimSpace(string(run("--rsync-arg=-v", "backup-server"))), "\n")
	if len(defaultArgs) < len(defaultRsyncOptions)+1 ||
		!reflect.DeepEqual(defaultArgs[:len(defaultRsyncOptions)], defaultRsyncOptions) ||
		defaultArgs[len(defaultRsyncOptions)] != "-v" {
		t.Fatalf("default rsync arguments = %q", defaultArgs)
	}
	if err := os.Mkdir(filepath.Join(dir, ".rsyncplan"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".rsyncplan", "config.txt"), []byte("--default-config\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, "custom.conf")
	if err := os.WriteFile(config, []byte("-a\n--exclude=/path with spaces\n"), 0600); err != nil {
		t.Fatal(err)
	}
	data := run(
		"--config="+config, "--rsync-arg=--no-owner", "--rsync-arg=-v",
		"--filesystem=/source/", "backup-server")
	args := strings.Split(strings.TrimSpace(string(data)), "\n")
	wantPrefix := []string{"-a", "--exclude=/path with spaces", "--no-owner", "-v",
		"--link-dest=/backups/example/rootfs/2025-01-01_000000.000000000/", "/source/"}
	if len(args) != len(wantPrefix)+1 || !reflect.DeepEqual(args[:len(wantPrefix)], wantPrefix) ||
		!strings.HasPrefix(args[len(args)-1], "backup-server:/backups/") {
		t.Fatalf("rsync arguments = %q", args)
	}

	for _, path := range []string{"", filepath.Join(dir, "missing")} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMainHelper$", "--", "--config="+path, "backup-server")
		cmd.Env = append(os.Environ(), "RSYNCPLAN_TEST_HELPER=1")
		if output, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("invalid config %q unexpectedly succeeded: %s", path, output)
		}
	}
}

func TestMainHelper(t *testing.T) {
	if os.Getenv("RSYNCPLAN_TEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"rsyncplan"}, os.Args[i+1:]...)
			flag.CommandLine = flag.NewFlagSet("rsyncplan", flag.ExitOnError)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing helper arguments")
}
