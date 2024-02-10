package template

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"reflect"
	"testing"
)

var testFolderName = "testdata"

func TestRenderTemplateValues(t *testing.T) {
	basicTplPath := path.Join(testFolderName, "basic.tpl")
	diffDelimsTplPath := path.Join(testFolderName, "diffDelims.tpl")
	unknownParamTplPath := path.Join(testFolderName, "unknownParameter.tpl")
	unknownFuncTplPath := path.Join(testFolderName, "unknownFunc.tpl")
	multiLevelParamsTplPath := path.Join(testFolderName, "multiLevelParams.tpl")
	dataSourcePath := path.Join(testFolderName, "templateDataSource.json")

	basicTplBytes, err := os.ReadFile(basicTplPath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", basicTplPath, err)
		return
	}

	diffDelimsTplBytes, err := os.ReadFile(diffDelimsTplPath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", basicTplPath, err)
		return
	}

	unknownParamTplBytes, err := os.ReadFile(unknownParamTplPath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", unknownParamTplPath, err)
		return
	}

	unknownFuncTplBytes, err := os.ReadFile(unknownFuncTplPath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", unknownFuncTplPath, err)
		return
	}

	multiLevelParamsTplBytes, err := os.ReadFile(multiLevelParamsTplPath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", multiLevelParamsTplPath, err)
		return
	}

	dataSourceBytes, err := os.ReadFile(dataSourcePath)
	if err != nil {
		t.Errorf("unable to read the '%s' file: %v", dataSourcePath, err)
		return
	}

	dataSource := &map[string]interface{}{}

	json.Unmarshal(dataSourceBytes, dataSource)

	ignoreOpts := NewRenderTemplateOptions()
	ignoreOpts.IgnoreTemplateErrors = true
	type args struct {
		templateContent string
		path            string
		dataSource      any
		replacements    []string
		ctx             context.Context
		opt             RenderTemplateOptions
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				templateContent: string(basicTplBytes),
				path:            basicTplPath,
				dataSource:      dataSource,
				replacements:    []string{},
				ctx:             context.Background(),
				opt:             RenderTemplateOptions{},
			},
			want: []byte("Hi Ken Cloutier"),
		},
		{
			name: "different delims",
			args: args{
				templateContent: string(diffDelimsTplBytes),
				path:            diffDelimsTplPath,
				dataSource:      dataSource,
				replacements:    []string{},
				ctx:             context.Background(),
				opt: RenderTemplateOptions{
					LeftDelim:  "||",
					RightDelim: "||",
				},
			},
			want: []byte("Hi Ken Cloutier"),
		},
		{
			name: "unknown parameter",
			args: args{
				templateContent: string(unknownParamTplBytes),
				path:            unknownParamTplPath,
				dataSource:      dataSource,
				replacements:    []string{},
				ctx:             context.Background(),
				opt:             NewRenderTemplateOptions(),
			},
			want:    []byte(""),
			wantErr: true,
		},
		{
			name: "unknown parameter ignore template errors",
			args: args{
				templateContent: string(unknownParamTplBytes),
				path:            unknownParamTplPath,
				dataSource:      dataSource,
				replacements:    []string{},
				ctx:             context.Background(),
				opt:             ignoreOpts,
			},
			want: []byte("Hi Ken Cloutier <no value>"),
		},
		{
			name: "unknown function should fail",
			args: args{
				templateContent: string(unknownFuncTplBytes),
				path:            unknownFuncTplPath,
				dataSource:      dataSource,
				replacements:    []string{},
				ctx:             context.Background(),
				opt:             NewRenderTemplateOptions(),
			},
			want:    []byte{},
			wantErr: true,
		},
		{
			name: "replacement variables",
			args: args{
				templateContent: string(multiLevelParamsTplBytes),
				path:            multiLevelParamsTplPath,
				dataSource:      dataSource,
				replacements:    []string{"different_and_conflicting_templating"},
				ctx:             context.Background(),
				opt:             NewRenderTemplateOptions(),
			},
			want: []byte("Hi Ken Cloutier {{ different_and_conflicting_templating }}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplateValues(tt.args.ctx, tt.args.templateContent, tt.args.path, tt.args.dataSource, tt.args.replacements, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplateValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RenderTemplateValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
