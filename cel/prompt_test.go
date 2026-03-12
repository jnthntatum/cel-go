// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cel

import (
	_ "embed"
	"sync"
	"testing"

	"github.com/google/cel-go/common"
	"github.com/google/cel-go/common/env"
	"github.com/google/go-cmp/cmp"

	"google.golang.org/protobuf/proto"
	dpb "google.golang.org/protobuf/types/descriptorpb"
)

//go:embed testdata/basic.prompt.txt
var wantBasicPrompt string

//go:embed testdata/macros.prompt.txt
var wantMacrosPrompt string

//go:embed testdata/standard_env.prompt.txt
var wantStandardEnvPrompt string

//go:embed testdata/field_paths.prompt.txt
var wantFieldPathsPrompt string

//go:embed testdata/test_fds_with_source_info-transitive-descriptor-set-source-info.proto.bin
var testFdsWithSourceInfo []byte

var onceFds sync.Once
var fds *dpb.FileDescriptorSet

func testFds(t *testing.T) *dpb.FileDescriptorSet {
	onceFds.Do(func() {
		fds = &dpb.FileDescriptorSet{}
		err := proto.Unmarshal(testFdsWithSourceInfo, fds)
		if err != nil {
			t.Fatalf("failed to unmarshal testFdsWithSourceInfo: %v", err)
		}
	})
	return fds
}

func TestPromptTemplate(t *testing.T) {
	tests := []struct {
		name    string
		envOpts []EnvOption
		out     string
	}{
		{
			name: "basic",
			out:  wantBasicPrompt,
		},
		{
			name:    "macros",
			envOpts: []EnvOption{Macros(StandardMacros...)},
			out:     wantMacrosPrompt,
		},
		{
			name:    "standard_env",
			envOpts: []EnvOption{StdLib(StdLibSubset(env.NewLibrarySubset().SetDisableMacros(true)))},
			out:     wantStandardEnvPrompt,
		},
	}

	for _, tst := range tests {
		tc := tst
		t.Run(tc.name, func(t *testing.T) {
			envOpts := append([]EnvOption{TypeDescs(testFds(t))}, tc.envOpts...)

			env, err := NewCustomEnv(
				envOpts...,
			)
			if err != nil {
				t.Fatalf("cel.NewCustomEnv() failed: %v", err)
			}
			prompt, err := AuthoringPrompt(env)
			if err != nil {
				t.Fatalf("cel.AuthoringPrompt() failed: %v", err)
			}
			out := prompt.Render("<USER_PROMPT>")
			if diff := cmp.Diff(tc.out, out); diff != "" {
				t.Errorf("got %s, diff (-want +got): %s", out, diff)
			}
		})
	}
}

func TestPromptTemplateFieldPaths(t *testing.T) {
	tests := []struct {
		name    string
		envOpts []EnvOption
		out     string
	}{
		{
			name: "standard_env",
			envOpts: []EnvOption{
				VariableWithDoc("team", ObjectType("cel.testdata.Team"),
					common.MultilineDescription("A team of gifted youngsters")),
				StdLib(StdLibSubset(env.NewLibrarySubset().SetDisableMacros(true))),
			},
			out: wantFieldPathsPrompt,
		},
	}

	for _, tst := range tests {
		tc := tst
		t.Run(tc.name, func(t *testing.T) {
			envOpts := append([]EnvOption{TypeDescs(testFds(t))}, tc.envOpts...)

			env, err := NewCustomEnv(
				envOpts...,
			)
			if err != nil {
				t.Fatalf("cel.NewCustomEnv() failed: %v", err)
			}
			prompt, err := AuthoringPromptWithFieldPaths(env)
			if err != nil {
				t.Fatalf("cel.AuthoringPromptWithFieldPaths() failed: %v", err)
			}
			out := prompt.Render("<USER_PROMPT>")
			if diff := cmp.Diff(tc.out, out); diff != "" {
				t.Errorf("got %s, diff (-want +got): %s", out, diff)
			}
		})
	}
}
