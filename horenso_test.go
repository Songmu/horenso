package horenso

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"
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
	defer func() {
		for _, f := range []string{noticeReport, fname, fname2} {
			os.RemoveAll(f)
		}
	}()
	_, ho, cmdArgs, err := parseArgs([]string{
		"--noticer",
		"go run testdata/reporter.go " + noticeReport,
		"-n", "invalid",
		"--reporter",
		"go run testdata/reporter.go " + fname,
		"-r",
		"go run testdata/reporter.go " + fname2,
		"--",
		"go", "run", "testdata/run.go",
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	ho.errStream = ioutil.Discard
	ho.outStream = ioutil.Discard

	r, err := ho.run(cmdArgs)
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}

	if r.ExitCode != 0 {
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
	expectedHostname, _ := os.Hostname()
	if r.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s but: %s", expectedHostname, r.Hostname)
	}

	rr := parseReport(fname)
	if !deepEqual(r, rr) {
		t.Errorf("something went wrong. expect: %#v, got: %#v", r, rr)
	}
	rr2 := parseReport(fname2)
	if !deepEqual(r, rr2) {
		t.Errorf("something went wrong. expect: %#v, got: %#v", r, rr2)
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
	if nr.ExitCode != -1 {
		t.Errorf("ExitCode should be -1")
	}
	if nr.Hostname != r.Hostname {
		t.Errorf("something went wrong")
	}
}

func TestRun_log(t *testing.T) {
	fname := temp()
	logfile := temp()
	defer func() {
		for _, f := range []string{fname, logfile} {
			os.RemoveAll(f)
		}
	}()
	_, ho, cmdArgs, err := parseArgs([]string{
		"--reporter",
		"go run testdata/reporter.go " + fname,
		"--log", logfile,
		"-vv",
		"-T",
		"--",
		"go", "run", "testdata/run.go",
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	ho.errStream = ioutil.Discard
	ho.outStream = ioutil.Discard

	r, err := ho.run(cmdArgs)
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	if r.ExitCode != 0 {
		t.Errorf("exit code should be 0 but: %d", r.ExitCode)
	}

	expect := "1\n2\n3\n"
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
	expectedHostname, _ := os.Hostname()
	if r.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s but: %s", expectedHostname, r.Hostname)
	}

	rr := parseReport(fname)
	if !deepEqual(r, rr) {
		t.Errorf("something went wrong. expect: %#v, got: %#v", r, rr)
	}

	// 2019-02-11T16:39:18.507451+09:00 1
	// 2019-02-11T16:39:18.507666+09:00 2
	// 2019-02-11T16:39:18.507668+09:00 3
	logdata, _ := ioutil.ReadFile(logfile)
	timeRegStr := `[0-9]{4}-[0-9]{2}-[0-9]{2}T` +
		`[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{6}` +
		`(?:Z|[-+][0-9]{2}:[0-9]{2}) `
	reg := regexp.MustCompile(`\A` +
		timeRegStr + "1\n" +
		timeRegStr + "2\n" +
		timeRegStr + "3\n" + `\z`)
	if !reg.Match(logdata) {
		t.Errorf("something went wrong.\n%s", string(logdata))
	}
}

func TestRun_notfound(t *testing.T) {
	fname := temp()
	defer os.RemoveAll(fname)
	_, ho, cmdArgs, err := parseArgs([]string{
		"--reporter",
		"go run testdata/reporter.go " + fname,
		"--",
		"testdata/notfound",
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	ho.errStream = ioutil.Discard
	ho.outStream = ioutil.Discard

	r, err := ho.run(cmdArgs)
	if err == nil {
		t.Errorf("err shouldn't be nil")
	}

	if r.ExitCode != -1 {
		t.Errorf("exit code should be -1 but: %d", r.ExitCode)
	}

	if r.StartAt == nil {
		t.Errorf("StartAt shouldn't be nil")
	}
	if r.EndAt != nil {
		t.Errorf("EtartAt should be nil")
	}
	expectedHostname, _ := os.Hostname()
	if r.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s but: %s", expectedHostname, r.Hostname)
	}

	rr := parseReport(fname)
	if !deepEqual(r, rr) {
		t.Errorf("something went wrong. expect: %#v, got: %#v", r, rr)
	}
}

func TestRun_config(t *testing.T) {
	_, ho, cmdArgs, err := parseArgs([]string{
		"-r", "hhh",
		"--config", "testdata/config.yaml",
		"--",
		"go", "run", "testdata/run.go",
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	ho.errStream = ioutil.Discard
	ho.outStream = ioutil.Discard

	r, err := ho.run(cmdArgs)
	if err != nil {
		t.Errorf("err should be nil, but: %s", err)
	}

	if !reflect.DeepEqual(ho.Reporter, []string{"hhh", "hoge", "fuga"}) {
		t.Errorf("something went wrong")
	}

	if r.ExitCode != 0 {
		t.Errorf("exit code should be 0 but: %d", r.ExitCode)
	}

	if r.StartAt == nil {
		t.Errorf("StartAt shouldn't be nil")
	}
	if r.EndAt == nil {
		t.Errorf("EndAt shouldn't be nil")
	}
	expectedHostname, _ := os.Hostname()
	if r.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s but: %s", expectedHostname, r.Hostname)
	}
}

func TestRunHugeOutput(t *testing.T) {
	fname := temp()
	defer os.RemoveAll(fname)
	_, ho, cmdArgs, err := parseArgs([]string{
		"--reporter",
		"go run testdata/reporter.go " + fname,
		"--",
		"go", "run", "testdata/run_hugeoutput.go",
	})
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}
	ho.errStream = ioutil.Discard
	ho.outStream = ioutil.Discard

	r, err := ho.run(cmdArgs)
	if err != nil {
		t.Errorf("err should be nil but: %s", err)
	}

	if r.ExitCode != 0 {
		t.Errorf("exit code should be 0 but: %d", r.ExitCode)
	}

	expect := 64*1024 + 1
	if len(r.Output) != expect {
		t.Errorf("output should be %d bytes but: %d bytes", expect, len(r.Output))
	}
	if len(r.Stdout) != expect {
		t.Errorf("output should be %d bytes but: %d bytes", expect, len(r.Stdout))
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
	expectedHostname, _ := os.Hostname()
	if r.Hostname != expectedHostname {
		t.Errorf("Hostname should be %s but: %s", expectedHostname, r.Hostname)
	}

	rr := parseReport(fname)
	if !deepEqual(r, rr) {
		t.Errorf("something went wrong. expect: %#v, got: %#v", r, rr)
	}
}

func equalTimePtr(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil || t2 == nil {
		return false
	}
	return (*t1).Equal(*t2)
}

func deepEqual(r1, r2 Report) bool {
	return r1.Command == r2.Command &&
		reflect.DeepEqual(r1.CommandArgs, r2.CommandArgs) &&
		r1.Tag == r2.Tag &&
		r1.Output == r2.Output &&
		r1.Stdout == r2.Stdout &&
		r1.Stderr == r2.Stderr &&
		r1.ExitCode == r2.ExitCode &&
		r1.Result == r2.Result &&
		r1.Pid == r2.Pid &&
		r1.Hostname == r2.Hostname &&
		r1.Signaled == r2.Signaled &&
		equalTimePtr(r1.StartAt, r2.StartAt) &&
		equalTimePtr(r1.EndAt, r2.EndAt)
}
