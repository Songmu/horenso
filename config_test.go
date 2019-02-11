package horenso

import (
	"reflect"
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
