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

package cmd_test

import (
	"io/ioutil"
	"log"
	"os"

	sidekick "cuelang.org/go/cmd/cue/cmd"
)

func ExampleRunSidewise() {

	// import (
	//     sidekick "cuelang.org/go/cmd/cue/cmd"
	// )

	func(){
		content := []byte("answer: 2*3*7")
		ioutil.WriteFile("sample.cue", content, 0644)
	}()

	err := sidekick.RunSidewise("eval", "sample.cue")
	if err != nil {
		log.Println("Error", err)
	}

	func() {
		os.Remove("sample.cue")
	}()

	// Output:
	// answer: 42
}
