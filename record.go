// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	//"crypto/md5"
	//"encoding/json"
	//"fmt"
	//"io"
	//"io/ioutil"
	//"log"
	//"os"
	//"sort"
	"time"
	"net/http"
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
	Hash       string                    `json:"hash"`
	Duration   time.Duration             `json:"duration"`
	Name       string                    `json:"name"`
	Components map[string]CheckoutRecord `json:"components"`
	Steps      []StepRecord              `json:"steps"`
}

// CheckoutRecord
type CheckoutRecord struct {
	Duration  time.Duration `json:"duration"`
	Status    BuildStatus   `json:"status"`
	Name      string        `json:"name"`
	Url       string        `json:"url"`
	Revision  string        `json:"revision"`
}

// StepRecord
type StepRecord struct {
	Duration  time.Duration `json:"duration"`
	Status    BuildStatus   `json:"status"`
	Directory string        `json:"directory"`
	Command   []string      `json:"command"`
}

func NewBuildRecord(p Project) *BuildRecord {
	var r BuildRecord
	r.Name = p.Name
        r.Components = map[string]CheckoutRecord{}
	for cname, component := range p.Components {
		var co CheckoutRecord
		co.Name = component.Name
		co.Url = component.Url
		co.Revision = "TBD"
		co.Status = BuildNotRun
		r.Components[cname] = co
	}
        r.Steps = []StepRecord{}
	for index := range p.Steps {
		r.Steps = append(r.Steps, StepRecord{Directory: p.Steps[index].Directory, Command: p.Steps[index].Command, Status: BuildNotRun})
	}
	return &r
}

func (r *BuildRecord) SetRevision(cname, revision string, duration time.Duration, status BuildStatus) {
	component := r.Components[cname]
	component.Revision = revision
	component.Duration = duration
	component.Status = status
	r.Components[cname] = component
}

func (r *BuildRecord) SetStatus(index int, status BuildStatus, duration time.Duration) {
	step := r.Steps[index]
	step.Status = status
	step.Duration = duration
	r.Steps[index] = step
}

func (r *BuildRecord) NewProject() *Project {
	var p Project
	p.Name = r.Name
        p.Components = map[string]Component{}
	for cname, component := range r.Components {
		var c Component
		c.Name = component.Name
		c.Url = component.Url
		c.Revision = component.Revision
		p.Components[cname] = c
	}
        p.Steps = []BuildStep{}
	for index := range r.Steps {
		p.Steps = append(p.Steps, BuildStep{Directory: r.Steps[index].Directory, Command: r.Steps[index].Command})
	}
	// FIXME Add handling of environment variables
	return &p
}

func (record *BuildRecord) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
                WriteHttpJson(w, record)
        } else if r.Method == "POST" {
                http.NotFound(w, r)
        } else {
                http.NotFound(w, r)
        }
}

