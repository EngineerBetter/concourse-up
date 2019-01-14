package yaml_test

import (
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/util/yaml"
	yamlenc "github.com/ghodss/yaml"
)

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		ops     string
		vars    map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "no vars, no ops",
			in:      "foo: bar",
			ops:     "",
			vars:    nil,
			want:    "foo: bar",
			wantErr: false,
		},
		{
			name: "var subsitution",
			in:   "foo: ((v))",
			ops:  "",
			vars: map[string]interface{}{
				"v": "bar",
			},
			want:    "foo: bar",
			wantErr: false,
		},
		{
			name: "compound var subsitution",
			in:   "foo: ((v))",
			ops:  "",
			vars: map[string]interface{}{
				"v": []string{"bar"},
			},
			want:    "foo: [bar]",
			wantErr: false,
		},
		{
			name: "simple operations",
			in:   "foo: bar",
			ops: `
- type: replace
  path: /baz?
  value: qoux
- type: replace
  path: /meta?
  value: syntatic`,
			vars:    nil,
			want:    "foo: bar\nbaz: qoux\nmeta: syntatic",
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := yaml.Interpolate(test.in, test.ops, test.vars)
			if !test.wantErr && err != nil {
				t.Fatalf("Interpolate(%s, %#v, %v) - unexpected error: %v", test.in, test.ops, test.vars, err)
			}
			if test.wantErr && err == nil {
				t.Fatalf("Interpolate(%s, %#v, %v) - expected error, got nil", test.in, test.ops, test.vars)
			}
			var want_, got_ interface{}
			err = yamlenc.Unmarshal([]byte(test.want), &want_)
			if err != nil {
				t.Fatal(err)
			}
			err = yamlenc.Unmarshal([]byte(got), &got_)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got_, want_) {
				t.Fatalf("Interpolate(%s, %#v, %v) = %s; wanted %s", test.in, test.ops, test.vars, got, test.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	dummyYAML := `
---
a:
  b: aValue
`
	type args struct {
		b    []byte
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "successful path test",
			args: args{
				b:    []byte(dummyYAML),
				path: "a/b",
			},
			want:    "aValue\n",
			wantErr: false,
		},
		{
			name: "unknown path test",
			args: args{
				b:    []byte(dummyYAML),
				path: "a/c",
			},
			wantErr: true,
		},
		{
			name: "empty path test",
			args: args{
				b:    []byte(dummyYAML),
				path: "",
			},
			wantErr: true,
		},
		{
			name: "empty template test",
			args: args{
				b:    []byte(""),
				path: "a/b",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Path(tt.args.b, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Path() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Path() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
