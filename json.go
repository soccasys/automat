// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder implements a software build server which supports
// continuous integration processes.
package builder

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

// Load a project from a JSON formatted file.
func LoadJson(data *interface{}, file string) error {
        content, err := ioutil.ReadFile(file)
        if err != nil {
                log.Println("Error:", err)
                return err
        }
        if err := json.Unmarshal(content, &data); err != nil {
                log.Println("Error:", err)
                return err
        }
        return err
}

// Save a project to a JSON formatter file.
func SaveJson(data interface{}, file string) {
        text, _ := json.MarshalIndent(data, "", "    ")
        if err := ioutil.WriteFile(file, text, 0664); err != nil {
                log.Panic("Error:", err)
        }
}

func WriteHttpJson(w http.ResponseWriter, content interface{}) {
	text, _ := json.MarshalIndent(content, "", "    ")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, "%s\n", text)
}
