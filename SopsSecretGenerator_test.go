// Copyright 2019-2020 Go About B.V. and contributors
// Licensed under the Apache License, Version 2.0.

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"go.mozilla.org/sops/v3/pgp"
)

const testkeyFingerprint = "2D2483DF73A3A0FAEE3C2A695BDC395360CE8FF4"
const testkeyCrypttext = `
-----BEGIN PGP MESSAGE-----

hQEMA6z+tHR/duVIAQf/VO9SoP3PcDg7iZzm5CUQ1jgF6BGrJqL5azjF/D74I83t
UalWAsewjd+xcKLBSnzr+uf90NcfIGpWx9u3IIyOZgIDkshp89jxz+mzj1MHsXM4
/gMBzRwFoynYKtdFs6u5MnB/e4tw0tMkxgqp8tPog1FEEyWmGSV6WOpL0SOzc2mV
PUA167We9+DLR8Yj9ABBEPpFDDq2bucOV9d3bkQS6+hWBbTPZMGcHvcupdDtALjv
7FMKftz9FAoPNDSR4RXkXQSkD1jdNfx5d+SYxNyxR2IWk1pigkBzKPlL9BHFCHQD
KCRtg392tGQ7WRY0s1J5GHHvnaDd+p0m3LWbUG28JtI7AZZSOue8gect+bw+Z5bO
sDRQLEMprueQOkhMr/JzgWRCV8JYxRFOqXjl7PRBjmgNPMu2GzRIB8D/+SU=
=7dAF
-----END PGP MESSAGE-----
`

// Test suite setup

func TestMain(m *testing.M) {
	err := setupGnuPG()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func setupGnuPG() error {
	err := os.Setenv("GNUPGHOME", "testdata")
	if err != nil {
		return err
	}
	// Check whether the test key was loaded correctly
	key := pgp.NewMasterKeyFromFingerprint(testkeyFingerprint)
	key.EncryptedKey = testkeyCrypttext
	_, err = key.Decrypt()
	if err != nil {
		return errors.Wrap(err, "PGP decryption check failed")
	}
	return nil
}

// Tests

func Test_GenerateKRMManifest(t *testing.T) {
	type args struct {
		rlFile    string
		itemIndex int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"SecretFromEnv",
			args{"testdata/krm-function-input.yaml", 0},
			strings.TrimLeft(dedent.Dedent(`
				apiVersion: v1
				kind: Secret
				metadata:
				  name: secret-from-env
				data:
				  VAR_ENV: dmFsX2Vudg==
			`), "\n"),
			false,
		},
		{
			"SecretFromFile",
			args{"testdata/krm-function-input.yaml", 1},
			strings.TrimLeft(dedent.Dedent(`
				apiVersion: v1
				kind: Secret
				metadata:
				  name: secret-from-file
				data:
				  file.txt: c2VjcmV0Cg==
			`), "\n"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				in, _ := ioutil.ReadFile(tt.args.rlFile)
				out, err := fn.Run(fn.ResourceListProcessorFunc(generateKRMManifest), in)
				if (err != nil) != tt.wantErr {
					t.Errorf("generateKRMManifest() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				rl, _ := fn.ParseResourceList(out)
				got := fmt.Sprint(rl.Items[tt.args.itemIndex])
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("generateKRMManifest() got = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func Test_ProcessSopsSecretGenerator(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"SopsSecretGenerator",
			args{"testdata/generator.yaml"},
			strings.TrimLeft(dedent.Dedent(`
				apiVersion: v1
				kind: Secret
				metadata:
				  name: secret
				data:
				  file.txt: c2VjcmV0Cg==
			`), "\n"),
			false,
		},
		{"InvalidEnvs", args{"testdata/generator-invalidenv.yaml"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sopsSecretGeneratorManifest, _ := ioutil.ReadFile(tt.args.fn)
			got, err := processSopsSecretGenerator(sopsSecretGeneratorManifest)
			if (err != nil) != tt.wantErr {
				t.Errorf("processSopsSecretGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processSopsSecretGenerator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateSecret(t *testing.T) {
	type args struct {
		sopsSecret SopsSecretGenerator
	}
	tests := []struct {
		name    string
		args    args
		want    Secret
		wantErr bool
	}{
		{
			"Normal",
			args{
				SopsSecretGenerator{
					TypeMeta: TypeMeta{
						APIVersion: "goabout/v1beta1",
						Kind:       "SopsSecretGenerator",
					},
					ObjectMeta: ObjectMeta{
						Name:        "secret",
						Namespace:   "default",
						Labels:      kvMap{"label": "value"},
						Annotations: kvMap{"annotation": "value"},
					},
					Behavior:    "merge",
					EnvSources:  []string{"testdata/vars.env"},
					FileSources: []string{"testdata/file.txt"},
					Type:        "Oblique",
				},
			},
			Secret{
				TypeMeta: TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: ObjectMeta{
					Name:      "secret",
					Namespace: "default",
					Labels:    kvMap{"label": "value"},
					Annotations: kvMap{
						"annotation":                         "value",
						"kustomize.config.k8s.io/needs-hash": "true",
						"kustomize.config.k8s.io/behavior":   "merge",
					},
				},
				Data: kvMap{"VAR_ENV": b64("val_env"), "file.txt": b64("secret\n")},
				Type: "Oblique",
			},
			false,
		},
		{
			"InvalidSources",
			args{
				SopsSecretGenerator{
					TypeMeta: TypeMeta{
						APIVersion: "goabout/v1beta1",
						Kind:       "SopsSecretGenerator",
					},
					ObjectMeta: ObjectMeta{
						Name: "secret",
					},
					FileSources: []string{"testdata/missing.txt"},
				},
			},
			Secret{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				got, err := generateSecret(tt.args.sopsSecret)
				if (err != nil) != tt.wantErr {
					t.Errorf("generateSecret() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("generateSecret() got = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func Test_readFile(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ExistingFile", args{"testdata/generator.yaml"}, false},
		{"MissingFile", args{"testdata/missing.yaml"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readFile(tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
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
		want    SopsSecretGenerator
		wantErr bool
	}{
		{"SopsSecretGenerator", args{"testdata/generator.yaml"}, ssg(nil, []string{"testdata/file.txt"}), false},
		{"SopsSecret", args{"testdata/generator-oldkind.yaml"}, ssg(nil, []string{"testdata/file.txt"}), false},
		{"NotYaml", args{"testdata/notyaml.txt"}, SopsSecretGenerator{}, true},
		{"WrongVersion", args{"testdata/generator-wrongversion.yaml"}, SopsSecretGenerator{}, true},
		{"WrongKind", args{"testdata/generator-wrongkind.yaml"}, SopsSecretGenerator{}, true},
		{"NoName", args{"testdata/generator-noname.yaml"}, SopsSecretGenerator{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sopsSecretGeneratorManifest, _ := ioutil.ReadFile(tt.args.fn)
			got, err := readInput(sopsSecretGeneratorManifest)
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
		input SopsSecretGenerator
	}
	tests := []struct {
		name    string
		args    args
		want    kvMap
		wantErr bool
	}{
		{"Input", args{ssg([]string{"testdata/vars.env"}, []string{"testdata/file.txt"})}, kvMap{"VAR_ENV": b64("val_env"), "file.txt": b64("secret\n")}, false},
		{"EnvsError", args{ssg([]string{"testdata/file.txt"}, []string{"testdata/file.txt"})}, nil, true},
		{"FilesError", args{ssg([]string{"testdata/vars.env"}, []string{"testdata/missing.txt"})}, nil, true},
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
		{"Yaml", args{"testdata/file.yaml"}, kvMap{"file.yaml": b64("var: secret\n")}, false},
		{"Json", args{"testdata/file.json"}, kvMap{"file.json": b64("{\n\t\"var\": \"secret\"\n}")}, false},
		{"Env", args{"testdata/file.env"}, kvMap{"file.env": b64("VAR=secret\n")}, false},
		{"Ini", args{"testdata/file.ini"}, kvMap{"file.ini": b64("[section]\nvar = secret\n\n")}, false},
		{"Binary", args{"testdata/file.txt"}, kvMap{"file.txt": b64("secret\n")}, false},
		{"BinaryRenamed", args{"renamed.txt=testdata/file.txt"}, kvMap{"renamed.txt": b64("secret\n")}, false},
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

// Test util functions

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func b(s string) []byte {
	return []byte(s)
}

func ssg(envSources []string, fileSources []string) SopsSecretGenerator {
	return SopsSecretGenerator{
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
