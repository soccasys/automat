// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	//"regexp"
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
