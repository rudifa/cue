// Copyright 2023 Rudolf Farkas
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

package cmd

// This file defines the Run functions that are used by an application to run cue commands.

// Example:

// main.go --------------------------------------------------------------------
// package main

// import (
// 	"log"
// 	sidekick "cuelang.org/go/cmd/cue/cmd"
// )

// func main() {
// 	err := sidekick.RunSidewise("eval", "sample.cue")
// 	if err != nil {
// 		log.Println("Error", err)
// 	}
// }

// go run . --------------------------------------------------------------------
// cue-run-demo-2 % go run .
// value: 42

// sample.cue ------------------------------------------------------------------
// { value: 21*2 }

// go.mod ----------------------------------------------------------------------
// module hello-sidekick
// go 1.20

// replace cuelang.org/go => github.com/rudifa/cue v0.5.0-rudifa
// require cuelang.org/go v0.5.0

// require (...)

import (
	"context"
)

// RunSidewise runs the given cue command with the arguments.
func RunSidewise(cmd string, args ...string) error {

	allArgs := []string{cmd}
	allArgs = append(allArgs, args...)

	return mainErr(context.Background(), allArgs)
}
