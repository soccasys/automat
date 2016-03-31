// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder implements a software build server which supports
// continuous integration processes.
package builder

import (
	//"crypto/md5"
	//"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	//"os"
	//"sort"
	//"time"
	"net/http"
	"regexp"
)

// Database
type Database struct {
	Root     string             `json:"root"`
	Projects map[string]*Project `json:"projects"`
	Env      map[string]string  `json:"env"`
}

func NewDatabase(root string) *Database {
	var db Database
	log.Printf("Loading database: %s", root)
	db.Root = root
	db.Projects = map[string]*Project{}
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/projects", root))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		// Only look at files.
		if !file.IsDir() {
			log.Printf("Loading project: %s", file.Name())
			p := NewProject()
			p.Load(fmt.Sprintf("%s/projects/%s", root, file.Name()))
			// FIXME Check that the name of the file and the name of
			//       the file and the name of the project are matching.
			db.Projects[p.Name] = p
			log.Printf("Project loaded: %s", file.Name())
		}
	}
	db.Env = map[string]string{}
	log.Printf("Database loaded: %s", root)
	return &db
}

func (db *Database) Build(project string) string {
	log.Printf("Build started: %s", project)
        p, present := db.Projects[project]
        if !present {
		log.Panic("Project not found")
	}
	// FIXME The build is synchronous for now
	id := p.Build(fmt.Sprintf("%s/builds/%s", db.Root, project))
	log.Printf("Build finished: %s", project)
	return id
}

func (db *Database) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        reProject := regexp.MustCompile(`^/projects/([-A-Za-z_0-9]+)$`)
        reBuild := regexp.MustCompile(`^/projects/([-A-Za-z_0-9]+)/build$`)
        if pMatches := reProject.FindSubmatch([]byte(path)); pMatches != nil {
                p, present := db.Projects[string(pMatches[1])]
                if !present {
                        /* Return a 404 if the project requested does not exist. */
                        http.NotFound(w, r)
                        return
                }
                //err := json.Unmarshal(jsonText, p)
                //if err != nil {
                //        http.Error(w, "500 internal error", http.StatusInternalServerError)
                //        return
                //}
		if r.Method == "GET" {
			p.ServeHTTP(w, r)
                        return
		} else if r.Method == "POST" {
			//jsonText, err := ioutil.ReadAll(r.Body)
			//if err != nil {
			//        http.Error(w, "500 internal error", http.StatusInternalServerError)
			//        return
			//}
			http.NotFound(w, r)
                        return
		} else {
			http.NotFound(w, r)
                        return
		}
        } else if pMatches := reBuild.FindSubmatch([]byte(path)); pMatches != nil {
                p, present := db.Projects[string(pMatches[1])]
                if !present {
                        /* Return a 404 if the project requested does not exist. */
                        http.NotFound(w, r)
                        return
                }
		if r.Method == "GET" {
			db.Build(p.Name)
			p.ServeHTTP(w, r)
                        return
		} else if r.Method == "POST" {
			http.NotFound(w, r)
                        return
		} else {
			http.NotFound(w, r)
                        return
		}
	} else {
		http.NotFound(w, r)
                return
	}
}
