package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func __LINEETC__() string {
	_, file, line, ok := runtime.Caller(1) // Get info about the caller of this function
	if !ok {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d ", filepath.Base(file), line)
}

func main() {
	label := flag.String("label", "rootfs", "Backup label used in remote path (/backups/<host>/<label>/)")
	filesystem := flag.String("filesystem", "/", "Local filesystem/path to back up with rsync")
	flag.Parse()

	argsWithoutProg := flag.Args()
	RSYNCPLAN_DESTINATION_HOST := ""
	if len(argsWithoutProg) > 0 {
		RSYNCPLAN_DESTINATION_HOST = argsWithoutProg[len(argsWithoutProg)-1]
	}

	// Configure the standard logger to write to the syslog.
	// We set the priority to LOG_NOTICE and the tag to "mygoapp".
	logwriter, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, os.Args[0])
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("%s%s %s", __LINEETC__(), "Failed to connect to syslog, falling back to stderr: ", err.Error())
	} else {
		log.SetOutput(logwriter)
	}

	if len(argsWithoutProg) == 0 {
		log.Fatalf("%s %s", __LINEETC__(), "Failure to provide target backup server (ssh+rsync+etc)?")
		os.Exit(1)
	}
	if *label == "" || strings.Contains(*label, "/") {
		log.Fatalf("%s invalid --label %q (must be a single path component)", __LINEETC__(), *label)
		os.Exit(1)
	}
	if *filesystem == "" {
		log.Fatalf("%s invalid --filesystem (cannot be empty)", __LINEETC__())
		os.Exit(1)
	}

	//	lastArg
	me, err := os.Hostname()
	if err != nil {
		log.Fatalf("%s Cannot find hostname %s", __LINEETC__(), err.Error())
		os.Exit(10)
	}
	log.Printf("%s I am %s", __LINEETC__(), me)

	// Get the current time
	t := time.Now()

	// Format a custom date and time string
	customFormat := "2006-01-02_150405.000000000"
	log.Printf("%s Custom format: %s", __LINEETC__(), t.Format(customFormat))
	now := time.Now()
	layout := "2006-01-02_150405.000000000"

	// Format the current time using the layout.
	formattedTime := now.Format(layout)
	if err != nil {
		log.Fatalf("Cannot format timestamp %s", err.Error())
		os.Exit(9)
	}

	OFS := "/backups/" + me + "/" + *label + "/"
	log.Printf("%s %s %s %s", __LINEETC__(), "Calling ssh", RSYNCPLAN_DESTINATION_HOST, "ls -1d ... ")
	cmd := exec.Command("ssh", RSYNCPLAN_DESTINATION_HOST,
		"ls -1d "+OFS+"????-??-??*/ | sort -nr | head -n 20")

	// Capture the output
	outputString, err := cmd.Output()
	if err != nil {
		log.Printf("%s failure: %v", __LINEETC__(), err)
		//can be blank. continue on...		os.Exit(76)
	}

	log.Printf("%s from remote stdout: %s", __LINEETC__(), outputString)
	ld := strings.Split(string(outputString), "\n")
	// --link-dest=

	ops := "--rsync-path=/usr/local/sbin/rsyncplan-exechook --timeout=1200 --exclude=/swapfile -iSalx"
	opsArray := strings.Split(ops, " ")
	rootfs := *filesystem

	log.Printf("%s Target destination directory calculated as: %s", __LINEETC__(), OFS+formattedTime+"/")
	allopts := []string{}
	// log.Printf("%s exec: %v", __LINEETC__(), allopts)

	allopts = append(allopts, opsArray...)
	//	log.Printf("%s opsArray... %v", __LINEETC__(), opsArray)

	for _, v := range ld {
		if len(v) != 0 {
			//			log.Printf("%s appending %s (len(v)==%d", __LINEETC__(), "--link-dest="+v, len(v))
			allopts = append(allopts, "--link-dest="+v)
		}
	}

	log.Printf("%s ld... %v (all the --link-dest= targets)", __LINEETC__(), allopts)

	allopts = append(allopts, rootfs, RSYNCPLAN_DESTINATION_HOST+":"+OFS+formattedTime+"/")
	log.Printf("%s finally %v", __LINEETC__(), allopts)

	cmd = exec.Command("rsync", allopts...)
	cmd.Stdout = os.Stdout // Direct stdout to the program's stdout
	cmd.Stderr = os.Stderr // Direct stderr to the program's stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		log.Fatalf("%s %s %s", __LINEETC__(), "rsync", err.Error())
		os.Exit(255)
	}
	// log.Printf("%s stdout was %s", __LINEETC__(), cmd.Stdout)
}
