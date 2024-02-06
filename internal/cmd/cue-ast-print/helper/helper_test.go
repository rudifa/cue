// Copyright 2024 The CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper_test

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"cuelang.org/go/cue/parser"
	"cuelang.org/go/internal/cmd/cue-ast-print/helper"
)

func TestDebugAstStr(t *testing.T) {
	input := `// 2567.slice.1.cue
b:[
    a[2:],   // third slice
]
a: [1,2,3,4,5]
`
	ast, err := parser.ParseFile("2567.slice.1.cue", input, parser.ParseComments)

	if err != nil {
		t.Errorf("parser.ParseFile returned error: %v", err)
	}

	got := helper.DebugAstTree(ast)
	got = strings.ReplaceAll(got, "\t", " ")

	if err != nil {
		t.Errorf("debugAstStr returned error: %v", err)
	}

	gotsum := md5sum(got)
	wantsum := "af4d735402e52fcdebb1b40adc158f39"

	if gotsum != wantsum {
		t.Errorf("debugAstStr got |%s|, want |%s|", gotsum, wantsum)
	}
}

func md5sum(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
