// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"bytes"
	"log"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runGitCheckout(url, name, ref, root string) (string, error) {
	// Check that the directory exists, if not create it.
	if _, errStat := os.Stat(root); os.IsNotExist(errStat) {
		if errMkdir := os.MkdirAll(root, 0755); errMkdir != nil {
			return ref, fmt.Errorf("Failed to create directory ", errMkdir)
		}
	}
	if _, errStat := os.Stat(root + "/" + name + "/.git"); os.IsNotExist(errStat) {
		cmd := exec.Command("git", "clone", url, name)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if errClone := cmd.Run(); errClone != nil {
			return ref, fmt.Errorf("Failed to clone ", errClone)
		}
	}
	cmd := exec.Command("git", "fetch")
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if errFetch := cmd.Run(); errFetch != nil {
		return ref, fmt.Errorf("Failed to fetch ", errFetch)
	}
	cmd = exec.Command("git", "clean", "-d", "-f", "-x")
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if errClean := cmd.Run(); errClean != nil {
		return ref, fmt.Errorf("Failed to clean ", errClean)
	}
	cmd = exec.Command("git", "checkout", ref)
	cmd.Dir = root + "/" + name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if errCo := cmd.Run(); errCo != nil {
		return ref, fmt.Errorf("Failed to checkout ", errCo)
	}
	return findGitHeadCommit(name, root)
}

func findGitHeadCommit(name, root string) (string, error) {
	if _, errStat := os.Stat(root + "/" + name + "/.git"); os.IsNotExist(errStat) {
		return "ERROR", fmt.Errorf("Git repository not found ", errStat)
	}
	buffer := bytes.Buffer{}
	cmd := exec.Command("git", "rev-list", "-n", "1", "HEAD")
	cmd.Dir = root + "/" + name
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "ERROR", fmt.Errorf("Failed to read HEAD Commit-ID ", err)
	}
	return strings.TrimSpace(buffer.String()), nil
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

func runCommand(directory, command string, args ...string) error {
	if _, errStat := os.Stat(directory); os.IsNotExist(errStat) {
		return fmt.Errorf("Directory not found %s", directory)
	}
	cmd := exec.Command(command, args...)
	cmd.Dir = directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// FIXME Add handling of environment here
	err := cmd.Run()
	return err
}
