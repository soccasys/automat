// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func writeJson(w http.ResponseWriter, content interface{}) {
	text, _ := json.MarshalIndent(content, "", "    ")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, "%s\n", text)
}

func (p *Project) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		writeJson(w, p)
	} else if r.Method == "POST" {
		http.NotFound(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (p *Project) postHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	jsonText, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "500 internal error", http.StatusInternalServerError)
		return
	}
	reProject := regexp.MustCompile(`^/project/([A-Za-z_0-9]+)$`)
	reBuild := regexp.MustCompile(`^/project/([A-Za-z_0-9]+)/build$`)
	if pMatches := reProject.FindSubmatch([]byte(path)); pMatches != nil {
		p, present := db.Projects[string(pMatches[1])]
		if !present {
			/* Return a 404 if the project requested does not exist. */
			http.NotFound(w, r)
			return
		}
		err := json.Unmarshal(jsonText, p)
		if err != nil {
			http.Error(w, "500 internal error", http.StatusInternalServerError)
			return
		}
		writeJson(w, p)
	} else if bMatches := reBuild.FindSubmatch([]byte(path)); bMatches != nil {
		p, present := db.Projects[string(bMatches[1])]
		if !present {
			/* Return a 404 if the project requested does not exist. */
			http.NotFound(w, r)
			return
		}
		//var build builder.BuildRequest
		//err := json.Unmarshal(jsonText, build)
		//if err != nil {
		//    http.Error(w, "500 internal error", http.StatusInternalServerError)
		//    return
		//}
		go func() { db.Build(string(bMatches[1])) }()
		writeJson(w, p)
		// TODO Start running a build for this project
	} else {
		/* Return a 404 if the project requested does not exist. */
		http.NotFound(w, r)
		return
	}
}

func (p *Project) getHandler(w http.ResponseWriter, r *http.Request) {
			writeJson(w, p)
		} else {
			/* Return a 404 if the project requested does not exist. */
			http.NotFound(w, r)
			return
		}
	}
}
