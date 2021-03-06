/**
* Copyright 2018 Comcast Cable Communications Management, LLC
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package gzip

import (
	"io/ioutil"
	"testing"
)

func TestInflate(t *testing.T) {
	const expected = "this is the inflated text string"
	c, err := ioutil.ReadFile("../../../../testdata/gzip_test.txt.gz")
	if err != nil {
		t.Error(err)
	}
	u, err := Inflate(c)
	if err != nil {
		t.Error(err)
	}
	if string(u) != expected {
		t.Errorf(`got "%s" expected "%s"`, string(u), expected)
	}

	_, err = Inflate(nil)
	if err == nil {
		t.Errorf("expected error: EOF")
	}

}
