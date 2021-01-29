// Copyright Dapr Authors
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

package config

import (
	"reflect"
	"testing"

	"github.com/dapr/dapr/dapr-agent/pkg/model"
)

func TestGetMeshConfig(t *testing.T) {

	cases := []struct {
		name        string
		annotation  string
		environment string
		file        string
		expect      model.NodeMetadata
	}{
		{
			name:   "Defaults",
			expect: *model.DefaultMetaConfig(),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMeshConfig()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("got \n%v expected \n%v", got, tt.expect)
			}
		})
	}
}
