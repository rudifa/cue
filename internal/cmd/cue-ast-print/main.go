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

// cue-ast-print parses a CUE file and prints its syntax tree, for example:
//
//	cue-ast-print file.cue
package main

import (
	"flag"
	"log"

	"fmt"

	"cuelang.org/go/cue/parser"
	"cuelang.org/go/internal/cmd/cue-ast-print/helper"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		// We could support multiple arguments or stdin if useful.
		log.Fatalln("cue-ast-print expects exactly one argument")
	}

	file, err := parser.ParseFile(args[0], nil, parser.ParseComments)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(helper.DebugAstTree(file))
}
