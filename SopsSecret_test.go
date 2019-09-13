// Copyright 2019 Go About B.V.
// Licensed under the Apache License, Version 2.0.

package main

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func b(s string) []byte {
	return []byte(s)
}

func ss(envSources []string, fileSources []string) SopsSecret {
	return SopsSecret{
		TypeMeta: TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		ObjectMeta: ObjectMeta{
			Name:        "secret",
			Annotations: kvMap{},
		},
		DisableNameSuffixHash: true,
		EnvSources:            envSources,
		FileSources:           fileSources,
	}
}

func Test_generateSecret(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateSecret(tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateSecret() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readInput(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name    string
		args    args
		want    SopsSecret
		wantErr bool
	}{
		{"SopsSecret", args{"testdata/sopssecret.yaml"}, ss(nil, []string{"testdata/file.txt"}), false},
		{"Missing", args{"testdata/missing.yaml"}, SopsSecret{}, true},
		{"NotYaml", args{"testdata/notyaml.txt"}, SopsSecret{}, true},
		{"WrongVersion", args{"testdata/sopssecret-wrongversion.yaml"}, SopsSecret{}, true},
		{"WrongKind", args{"testdata/sopssecret-wrongkind.yaml"}, SopsSecret{}, true},
		{"NoName", args{"testdata/sopssecret-noname.yaml"}, SopsSecret{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readInput(tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("readInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readInput() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseInput(t *testing.T) {
	type args struct {
		input SopsSecret
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Input", args{ss([]string{"testdata/vars.env"}, []string{"testdata/file.txt"})}, kvMap{"VAR_ENV": b64("val_env"), "file.txt": b64("secret\n")}, false},
		{"EnvsError", args{ss([]string{"testdata/file.txt"}, []string{"testdata/file.txt"})}, nil, true},
		{"FilesError", args{ss([]string{"testdata/vars.env"}, []string{"testdata/missing.txt"})}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInput() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEnvSources(t *testing.T) {
	type args struct {
		sources []string
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Envs", args{[]string{"testdata/vars.env", "testdata/vars.yaml"}}, kvMap{"VAR_ENV": b64("val_env"), "VAR_YAML": b64("val_yaml")}, false},
		{"NoEnvs", args{[]string{}}, kvMap{}, false},
		{"Error", args{[]string{"testdata/missing.env"}}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseEnvSources(tt.args.sources, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEnvSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnvSources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEnvSource(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"DotEnv", args{"testdata/vars.env"}, kvMap{"VAR_ENV": b64("val_env")}, false},
		{"YAML", args{"testdata/vars.yaml"}, kvMap{"VAR_YAML": b64("val_yaml")}, false},
		{"JSON", args{"testdata/vars.json"}, kvMap{"VAR_JSON": b64("val_json")}, false},
		{"Binary", args{"testdata/file.txt"}, kvMap{}, true},
		{"Missing", args{"testdata/missing.txt"}, kvMap{}, true},
		{"NotSops", args{"testdata/empty.txt"}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseEnvSource(tt.args.source, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEnvSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnvSource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDotEnvContent(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Variables", args{b("VAR1=val1\nVAR2=val2")}, kvMap{"VAR1": b64("val1"), "VAR2": b64("val2")}, false},
		{"StringBOM", args{append(utf8bom, b("VAR=val")...)}, kvMap{"VAR": b64("val")}, false},
		{"Empty", args{b("")}, kvMap{}, false},
		{"InvalidLine", args{b("VAR")}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseDotEnvContent(tt.args.content, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDotEnvContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDotEnvContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDotEnvLine(t *testing.T) {
	type args struct {
		line []byte
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Variable", args{b("VAR=value")}, kvMap{"VAR": b64("value")}, false},
		{"TrimLeft", args{b(" VAR=value")}, kvMap{"VAR": b64("value")}, false},
		{"EmptyLine", args{b("")}, kvMap{}, false},
		{"Comment", args{b("# Comment")}, kvMap{}, false},
		{"NoValue", args{b("VAR")}, kvMap{}, true},
		{"InvalidUTF8", args{[]byte{0xff, 0xfe, 0xfd}}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseDotEnvLine(tt.args.line, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDotEnvLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDotEnvLine() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseYAMLContent(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Variables", args{b("VAR1: val1\nVAR2: val2")}, kvMap{"VAR1": b64("val1"), "VAR2": b64("val2")}, false},
		{"Empty", args{b("")}, kvMap{}, false},
		{"InvalidSyntax", args{b("VAR:val")}, kvMap{}, true},
		{"InvalidType", args{b("VAR: [1, 2]")}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseYAMLContent(tt.args.content, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseYAMLContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseYAMLContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseJsonContent(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Variables", args{b(`{"VAR1": "val1", "VAR2": "val2"}`)}, kvMap{"VAR1": b64("val1"), "VAR2": b64("val2")}, false},
		{"Empty", args{b(`{}`)}, kvMap{}, false},
		{"InvalidSyntax", args{b(`{"VAR"}`)}, kvMap{}, true},
		{"InvalidType", args{b(`{"VAR": ["val"]}`)}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseJSONContent(tt.args.content, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJSONContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFileSources(t *testing.T) {
	type args struct {
		sources []string
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Files", args{[]string{"testdata/file.txt", "testdata/file2.txt"}}, kvMap{"file.txt": b64("secret\n"), "file2.txt": b64("secret2\n")}, false},
		{"NoFiles", args{[]string{}}, kvMap{}, false},
		{"Error", args{[]string{"testdata/missing.txt"}}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseFileSources(tt.args.sources, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileSources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFileSources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFileSource(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"File", args{"testdata/file.txt"}, kvMap{"file.txt": b64("secret\n")}, false},
		{"MissingFile", args{"testdata/missing.txt"}, kvMap{}, true},
		{"InvalidName", args{"=testdata/file.txt"}, kvMap{}, true},
		{"NotSopsFile", args{"testdata/empty.txt"}, kvMap{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(kvMap)
			err := parseFileSource(tt.args.source, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFileSource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFileName(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name     string
		args     args
		wantKey  string
		wantFile string
		wantErr  bool
	}{
		{"WithoutDirectory", args{"filename"}, "filename", "filename", false},
		{"WithDirectory", args{"directory/filename"}, "filename", "directory/filename", false},
		{"ExplicitKey", args{"key=filename"}, "key", "filename", false},
		{"MissingKey", args{"=filename"}, "", "", true},
		{"MissingFilename", args{"key="}, "", "", true},
		{"TooManyEqualSigns", args{"key=filename=extra"}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseFileName(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantKey {
				t.Errorf("parseFileName() got = %v, wantKey %v", got, tt.wantKey)
			}
			if got1 != tt.wantFile {
				t.Errorf("parseFileName() got1 = %v, wantFile %v", got1, tt.wantFile)
			}
		})
	}
}

func Test_formatForPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"YAMLLong", args{"dir/file.yaml"}, "yaml"},
		{"YAMLShort", args{"dir/file.yml"}, "yaml"},
		{"JSON", args{"dir/file.json"}, "json"},
		{"DotEnv", args{"dir/file.env"}, "dotenv"},
		{"Other", args{"dir/file.txt"}, "binary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatForPath(tt.args.path); got != tt.want {
				t.Errorf("formatForPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
