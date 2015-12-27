package horenso

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func temp() string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return f.Name()
}

func parseReport(fname string) Report {
	byt, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("failed to read " + fname)
	}
	r := Report{}
	json.Unmarshal(byt, &r)
	return r
}

func TestRun(t *testing.T) {
	noticeReport := temp()
	fname := temp()
	fname2 := temp()
	o, err := parseArgs([]string{
		"--noticer",
		"go run _testdata/reporter.go " + noticeReport,
		"-n", "invalid",
		"--reporter",
		"go run _testdata/reporter.go " + fname,
		"-r",
		"go run _testdata/reporter.go " + fname2,
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	r, err := o.run([]string{"go", "run", "_testdata/run.go"})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}

	if *r.ExitCode != 0 {
		t.Errorf("exit code should be 0 but: %d", r.ExitCode)
	}

	expect := "1\n2\n3\n"
	if r.Output != expect {
		t.Errorf("output should be %s but: %s", expect, r.Output)
	}
	if r.Stdout != expect {
		t.Errorf("output should be %s but: %s", expect, r.Stdout)
	}
	if r.Stderr != "" {
		t.Errorf("output should be empty but: %s", r.Stderr)
	}
	if r.StartAt == nil {
		t.Errorf("StartAt shouldn't be nil")
	}
	if r.EndAt == nil {
		t.Errorf("EtartAt shouldn't be nil")
	}

	rr := parseReport(fname)
	if !reflect.DeepEqual(r, rr) {
		t.Errorf("something went wrong")
	}
	rr2 := parseReport(fname2)
	if !reflect.DeepEqual(r, rr2) {
		t.Errorf("something went wrong")
	}

	nr := parseReport(noticeReport)
	if nr.Pid != r.Pid {
		t.Errorf("something went wrong")
	}
	if nr.Output != "" {
		t.Errorf("something went wrong")
	}
	if nr.StartAt == nil {
		t.Errorf("StartAt shouldn't be nil")
	}
	if nr.EndAt != nil {
		t.Errorf("EndAt should be nil")
	}
	if nr.ExitCode != nil {
		t.Errorf("ExitCode should be nil")
	}
}
