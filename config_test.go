package horenso

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c, err := loadConfig("testdata/config.yaml")
	if err != nil {
		t.Errorf("failed to load config: %s", err)
	}
	expect := config{
		Reporter: []string{"hoge", "fuga"},
		Noticer:  []string{"bar"},
	}
	if !reflect.DeepEqual(expect, *c) {
		t.Errorf("something went wrong\n   got: %#v\nexpect: %#v", *c, expect)
	}
}

func TestLoadConfigHandlers(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    handlers
		wantErr bool
	}{
		{name: "string", input: "echo report", want: handlers{"echo report"}},
		{name: "sequence", input: "[echo report, echo another]", want: handlers{"echo report", "echo another"}},
		{name: "empty sequence", input: "[]", want: handlers{}},
		{name: "number", input: "42", wantErr: true},
		{name: "mixed sequence", input: "[echo report, 42]", wantErr: true},
		{name: "mapping", input: "{command: echo report}", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(file, []byte("reporter: "+tt.input+"\nnoticer: "+tt.input+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := loadConfig(file)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "handlers should be a string or an array of string") {
					t.Fatalf("expected handler validation error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Reporter, tt.want) || !reflect.DeepEqual(got.Noticer, tt.want) {
				t.Errorf("unexpected handlers: reporter=%#v, noticer=%#v; want %#v", got.Reporter, got.Noticer, tt.want)
			}
		})
	}
}
