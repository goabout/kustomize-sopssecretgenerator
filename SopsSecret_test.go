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
				t.Errorf("formatForPath() = %v, wantKey %v", got, tt.want)
			}
		})
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

func Test_parseEnvSource(t *testing.T) {
	type args struct {
		source string
		data   kvMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseEnvSource(tt.args.source, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("parseEnvSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseEnvSources(t *testing.T) {
	type args struct {
		sources []string
		data    kvMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseEnvSources(tt.args.sources, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("parseEnvSources() error = %v, wantErr %v", err, tt.wantErr)
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
		{"ImplicitKey", args{"filename"}, "filename", "filename", false},
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

func Test_parseFileSource(t *testing.T) {
	type args struct {
		source string
		data   kvMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseFileSource(tt.args.source, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("parseFileSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseFileSources(t *testing.T) {
	type args struct {
		sources []string
		data    kvMap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseFileSources(tt.args.sources, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("parseFileSources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseInput(t *testing.T) {
	type args struct {
		input sopsSecret
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInput(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInput() got = %v, wantKey %v", got, tt.want)
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

func Test_readInput(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name    string
		args    args
		want    sopsSecret
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readInput(tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("readInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readInput() got = %v, wantKey %v", got, tt.want)
			}
		})
	}
}
