package util

import (
	"reflect"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	type args struct {
		name        string
		templateStr string
		params      interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Happy path of rendering a template",
			args: args{
				name:        "aTemplate",
				templateStr: `I will {{ .ReplaceThis }} with This`,
				params: struct {
					ReplaceThis string
				}{
					"This",
				},
			},
			want:    []byte("I will This with This"),
			wantErr: false,
		},
		{
			name: "Template rendering fails due to missing interpolation variable",
			args: args{
				name:        "aTemplate",
				templateStr: `I will {{ .ReplaceThis }} with This`,
				params: struct {
					CannotReplaceThis string
				}{
					"This",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.args.name, tt.args.templateStr, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
