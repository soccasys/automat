// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"bytes"
	"log"
	"time"
	"os"
	"os/exec"
	"strings"
)

func runGitCheckout(url, name, ref, root string) string {
	// Check that the directory exists, if not create it.
	if _, errStat := os.Stat(root); os.IsNotExist(errStat) {
		if errMkdir := os.MkdirAll(root, 0755); errMkdir != nil {
			log.Panic("Failed to create directory", errMkdir)
		}
	}
	if _, errStat := os.Stat(root + "/" + name + "/.git"); os.IsNotExist(errStat) {
		cmd := exec.Command("git", "clone", url, name)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Panic("Failed to clone", err)
		}
	}
	cmd := exec.Command("git", "fetch")
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to fetch", err)
	}
	cmd = exec.Command("git", "clean", "-d", "-f", "-x")
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to clean", err)
	}
	cmd = exec.Command("git", "checkout", ref)
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to checkout", err)
	}
	return findGitHeadCommit(name, root)
}

func findGitHeadCommit(name, root string) string {
	if _, errStat := os.Stat(root + "/" + name + "/.git"); os.IsNotExist(errStat) {
		log.Panic("Git repository not found", errStat)
	}
	buffer := bytes.Buffer{}
	cmd := exec.Command("git", "rev-list", "-n", "1", "HEAD")
	cmd.Dir = root + "/" + name
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to read HEAD Commit-ID", err)
	}
	return strings.TrimSpace(buffer.String())
}

func findGitRemoteCommit(name, ref, root string) string {
	if _, errStat := os.Stat(root + "/" + name + "/.git"); os.IsNotExist(errStat) {
		log.Panic("Git repository not found", errStat)
	}
	buffer := bytes.Buffer{}
	cmd := exec.Command("git", "rev-list", "-n", "1", "origin/"+ref)
	cmd.Dir = root + "/" + name
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to read remote Commit-ID", err)
	}
	return strings.TrimSpace(buffer.String())
}

func runCommand(name string, root string, command string, args ...string) (BuildStatus, time.Duration) {
	if _, errStat := os.Stat(root + "/" + name ); os.IsNotExist(errStat) {
		log.Panic("Directory not found: ", errStat)
	}
	cmd := exec.Command(command, args...)
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	start := time.Now()
	if err := cmd.Run(); err != nil {
		log.Panic("Failed to run command: ", err)
	}
	end := time.Now()
	return BuildOk, end.Sub(start)
}
