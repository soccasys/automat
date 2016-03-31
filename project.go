// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder implements a software build server which supports
// continuous integration processes.
package builder

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

type BuildStatus string

const (
	BuildNotRun  BuildStatus = "BUILD_NOT_RUN"
	BuildOk      BuildStatus = "BUILD_OK"
	BuildFailed  BuildStatus = "BUILD_FAILED"
	BuildSkipped BuildStatus = "BUILD_SKIPPED"
)

// BuildRecord is used to record the details of a build that has been run.
type BuildRecord struct {
	Hash       string
	Name       string
	Components map[string]Component
	Steps      []StepRecord
}

type StepRecord struct {
	Time      time.Duration
	Status    BuildStatus
	Directory string   `json:"directory"`
	Command   []string `json:"command"`
}

// Project
type Project struct {
	Name       string               `json:"name"`
	Components map[string]Component `json:"components"`
	Steps      []BuildStep          `json:"steps"`
	Env        map[string]string     `json:"env"`
}

type BuildStep struct {
	Description string   `json:"description"`
	Directory   string   `json:"directory"`
	Command     []string `json:"command"`
	Env         map[string]string `json:"env"`
}

type Component struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Revision string `json:"revision"`
}

func NewProject() *Project {
	var p Project
	p.Name = ""
	p.Components = map[string]Component{}
	p.Steps = []BuildStep{}
	p.Env = map[string]string{}
	return &p
}

func (p *Project) AddComponent(name, url, revision string) {
        var c Component
        c.Name = name
        c.Url = url
        c.Revision = revision
        p.Components[c.Name] = c
}

func (p *Project) AddBuildStep(description, directory string, command []string) {
        var step BuildStep
        step.Description = description
        step.Directory = directory
        step.Command = command
	step.Env = map[string]string{}
        p.Steps = append(p.Steps, step)
}

// Load a project from a JSON formatted file.
func (p *Project) Load(file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Panic("Error:", err)
	}
	if err := json.Unmarshal(content, &p); err != nil {
		log.Panic("Error:", err)
	}
}

// Save a project to a JSON formatter file.
func (p *Project) Save(file string) {
        text, _ := json.MarshalIndent(p, "", "    ")
        if err := ioutil.WriteFile(file, text, 0664); err != nil {
		log.Panic("Error:", err)
        }
}

// Build runs all the steps required to build a project, including first
// making sure all the GIT repositories are cloned and up-to-date, and
// then running through each of the build steps.
func (p *Project) Build(root string) string {
	log.Println("Build started")
	hash := md5.New()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Build failed", r)
		}
	}()
	// Create the root directory for the build if it does not exist yet.
	if _, errStat := os.Stat(root); os.IsNotExist(errStat) {
		if errMkdir := os.MkdirAll(root, 0755); errMkdir != nil {
			log.Panic("Failed to create build directory", errMkdir)
		}
	}
	io.WriteString(hash, "Components:\n")
	componentNames := make([]string, len(p.Components))
	i := 0
	for componentName, _ := range p.Components {
		componentNames[i] = componentName
		i += 1
	}
	sort.Strings(componentNames)
	for index := range componentNames {
		component := p.Components[componentNames[index]]
		// TODO record the commit ID in the build result
		commit := runGitCheckout(component.Url, component.Name, component.Revision, root)
		io.WriteString(hash, fmt.Sprintf("%s %s %s\n", component.Url, component.Name, commit))
	}
	io.WriteString(hash, "Steps:\n")
	for _, step := range p.Steps {
		runCommand(step.Directory, root, step.Command[0], step.Command[1:]...)
		io.WriteString(hash, step.Directory)
		for i, arg := range step.Command {
			io.WriteString(hash, fmt.Sprintf("%d: %s\n", i, arg))
		}
		io.WriteString(hash, "\n")
	}
	fmt.Printf("Hash: %x\n", hash.Sum(nil))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
