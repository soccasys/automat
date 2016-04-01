// Copyright (c) 2016, Socca Systems -- All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builder implements a software build server which supports
// continuous integration processes.
package builder

import (
	"fmt"
	"encoding/json"
	"net/http"
)

func writeJson(w http.ResponseWriter, content interface{}) {
	text, _ := json.MarshalIndent(content, "", "    ")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, "%s\n", text)
}
