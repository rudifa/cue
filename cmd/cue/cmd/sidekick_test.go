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
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	sidekick "cuelang.org/go/cmd/cue/cmd"
)

func ExampleRunSidewise() {

	// import (
	//     sidekick "cuelang.org/go/cmd/cue/cmd"
	// )

	func() {
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

func TestRunSidewise(t *testing.T) {

	const file = "sidekick-sample.cue"

	writeSampleFile(file)

	out, err := captureStdout(sidekick.RunSidewise, "eval", file)
	if err != nil {
		log.Println("Error", err)
	}

	removeSampleFile(file)

	expected := "answer: 42\n"
	if out != expected {
		t.Errorf("expected '%s', got '%s'", expected, out)
	}
}

func writeSampleFile(file string) {
	content := []byte("answer: 2*3*7")
	ioutil.WriteFile(file, content, 0644)
}

func removeSampleFile(file string) {
	os.Remove(file)
}

func captureStdout(f func(string, ...string) error, arg1, arg2 string) (string, error) {

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// call function
	err := f(arg1, arg2)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}
