// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package automat implements a software build server which supports
// continuous integration processes.
package automat

import (
	//"crypto/md5"
	//"encoding/json"
	//"io"
	"os"
	//"sort"
	//"time"
	"fmt"
	"io/ioutil"
	"log"
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
			err := p.Load(fmt.Sprintf("%s/projects/%s", root, file.Name()))
			// FIXME Check that the name of the file and the name of
			//       the file and the name of the project are matching.
			if err != nil {
				log.Printf("Project failed: %s %s", file.Name(), err)
			} else {
				db.Projects[p.Name] = p
				log.Printf("Project loaded: %s", file.Name())
			}
		}
	}
	db.Env = map[string]string{}
	log.Printf("Database loaded: %s", root)
	return &db
}

func (db *Database) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
	log.Printf("%s %s", r.Method, path)
        reProject := regexp.MustCompile(`^/projects/([-A-Za-z_0-9]+)$`)
        reBuild := regexp.MustCompile(`^/projects/([-A-Za-z_0-9]+)/build$`)
        if pMatches := reProject.FindSubmatch([]byte(path)); pMatches != nil {
                p, present := db.Projects[string(pMatches[1])]
                if !present && (r.Method == http.MethodGet || r.Method == http.MethodDelete) {
                        /* Return a 404 if the project requested does not exist. */
                        http.NotFound(w, r)
                        return
                }
		if r.Method == http.MethodGet {
			p.ServeHTTP(w, r)
                        return
		} else if r.Method == http.MethodDelete {
			// FIXME Better implementation is required here. Error checking for file removal, etc...
			p.ServeHTTP(w, r)
			delete(db.Projects, p.Name)
			os.Remove(fmt.Sprintf("%s/projects/%s", db.Root, p.Name))
                        return
		} else if r.Method == http.MethodPut || r.Method == http.MethodPost {
			// Create a new project, save it.
			p := NewProject()
			err := p.ReadJson(r.Body)
			if err != nil || p.Name != string(pMatches[1]) {
				http.Error(w, "400 bad request", http.StatusBadRequest)
				return
			}
			errSave := p.Save(fmt.Sprintf("%s/projects/%s", db.Root, p.Name))
			if errSave != nil {
				http.Error(w, "500 failed to save the project", http.StatusInternalServerError)
				return
			}
			db.Projects[p.Name] = p
			p.ServeHTTP(w, r)
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
	                // Create the root directory for the build if it does not exist yet.
	                root := fmt.Sprintf("%s/builds/%s", db.Root, p.Name)
	                if _, errStat := os.Stat(root); os.IsNotExist(errStat) {
		                if errMkdir := os.MkdirAll(root, 0755); errMkdir != nil {
					http.Error(w, "500 failed to create build directory", http.StatusInternalServerError)
					return
		                }
	                }
			record, errBuild := p.Build(root)
			if errBuild != nil {
				http.Error(w, "500 A critical error occured during the build", http.StatusInternalServerError)
				return
			}
			record.ServeHTTP(w, r)
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
