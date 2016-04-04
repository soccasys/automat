// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder implements a software build server which supports
// continuous integration processes.
package automat

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"io/ioutil"
	"log"
	"sort"
	"time"
	"strings"
	"net/http"
)

// Project
type Project struct {
	Name       string               `json:"name"`
	Components map[string]Component `json:"components"`
	Steps      []BuildStep          `json:"steps"`
	Env        map[string]string     `json:"env"`
}

// BuildStep
type BuildStep struct {
	Description string   `json:"description"`
	Directory   string   `json:"directory"`
	Command     []string `json:"command"`
	Env         map[string]string `json:"env"`
}

// Component
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
func (p *Project) Load(file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("Error:", err)
		return err
	}
	if err := json.Unmarshal(content, &p); err != nil {
		log.Println("Error:", err)
		return err
	}
	return err
}

// Load a project from a JSON formatted Reader
func (p *Project) ReadJson(r io.Reader) error {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		log.Println("Error:", err)
		return err
	}
	if err := json.Unmarshal(content, &p); err != nil {
		log.Println("Error:", err)
		return err
	}
	if p.Name == "" {
		return fmt.Errorf("Invalid Project Name")
	}
	return err
}

// Save a project to a JSON formatter file.
func (p *Project) Save(file string) error {
        text, _ := json.MarshalIndent(p, "", "    ")
        return ioutil.WriteFile(file, text, 0664)
}

// Build runs all the steps required to build a project:
// - Checkout all the components
// - Run all the build steps, with the project and step-level
//   environment variables set
//
// When an error occurs during the checkout of a component the
// other checkouts are attempted, but the build steps are skipped.
//
// When an error occurs during a build step, the following steps
// are skipped.
//
// The "root" directory must have been created prior to starting
// the build.
//
func (p *Project) Build(root string) (*BuildRecord, error) {
	log.Println("Build started")
	hash := md5.New()
	defer func() {
		if r := recover(); r != nil {
			log.Println("Build failed: ", r)
		}
	}()
	record := NewBuildRecord(*p)
	buildStart := time.Now()
	buildFailed := false
	io.WriteString(hash, "Components:\n")
	componentNames := make([]string, len(p.Components))
	i := 0
	for componentName, _ := range p.Components {
		componentNames[i] = componentName
		i += 1
	}
	log.Println("Git Checkout")
	sort.Strings(componentNames)
	for index := range componentNames {
		component := p.Components[componentNames[index]]
		start := time.Now()
		commit, err := runGitCheckout(component.Url, component.Name, component.Revision, root)
		end := time.Now()
		if err != nil {
			buildFailed = true
			record.SetRevision(componentNames[index], commit, end.Sub(start), BuildFailed)
		} else {
			record.SetRevision(componentNames[index], commit, end.Sub(start), BuildOk)
		}
		io.WriteString(hash, fmt.Sprintf("%s %s %s\n", component.Url, component.Name, commit))
	}
	log.Println("Build Steps")
	io.WriteString(hash, "Steps:\n")
	for index, step := range p.Steps {
		start := time.Now()
		if !buildFailed {
			directory := fmt.Sprintf("%s/%s", root, step.Directory)
			// Compute the environment to be passed to the command here.
			environment := ExpandEnvironment(root, p.Env, step.Env)
			err := runCommand(directory, environment, step.Command[0], step.Command[1:]...)
			end := time.Now()
			if err != nil {
				buildFailed = true
				record.SetStatus(index, BuildFailed, end.Sub(start))
			} else {
				record.SetStatus(index, BuildOk, end.Sub(start))
			}
		} else {
			// Skip build steps if a failure already occured.
			end := time.Now()
			record.SetStatus(index, BuildSkipped, end.Sub(start))
		}
		io.WriteString(hash, step.Directory)
		for i, arg := range step.Command {
			io.WriteString(hash, fmt.Sprintf("%d: %s\n", i, arg))
		}
		io.WriteString(hash, "\n")
	}
	log.Printf("Hash: %x\n", hash.Sum(nil))
	record.Hash = fmt.Sprintf("%x", hash.Sum(nil))
	buildEnd := time.Now()
	record.Duration = buildEnd.Sub(buildStart)
	return record, nil
}

func ExpandEnvironment (root string, projectEnv map[string]string, stepEnv map[string]string) []string {
	workEnv := map[string]string{}

	stepGetenv := func (name string) string {
		if name == "BUILD_ROOT" {
			return root
		}
		value, present := workEnv[name]
		if !present {
			return ""
		}
		return value
	}

	// Fill up the work environment starting with the current process environment.
	for _, key := range os.Environ() {
		nameValue := strings.SplitN(key, "=", 2)
		workEnv[nameValue[0]] = nameValue[1]
	}

	// Process Project environment variables
	for name := range projectEnv {
		workEnv[name] = os.Expand(projectEnv[name], stepGetenv)
	}
	// Process Step environment variables
	for name := range stepEnv {
		workEnv[name] = os.Expand(stepEnv[name], stepGetenv)
	}

	environment := []string{}
	for name := range workEnv {
		environment = append(environment, fmt.Sprintf("%s=%s", name, workEnv[name]))
	}
	return environment
}

func (p *Project) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		WriteHttpJson(w, p)
	} else if r.Method == http.MethodPut || r.Method == http.MethodPost {
		// Adding/Updating the project is handled by the database
		WriteHttpJson(w, p)
	} else {
		http.NotFound(w, r)
	}
}
