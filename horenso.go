package horenso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
)

const version = "0.0.1"

type opts struct {
	Reporter  []string `short:"r" long:"reporter" required:"true" value-name:"/path/to/reporter.pl" description:"handler for reporting the result of the job"`
	Noticer   []string `short:"n" long:"noticer" value-name:"/path/to/noticer.rb" description:"handler for noticing the start of the job"`
	TimeStamp bool     `short:"T" long:"timestamp" description:"add timestamp to merged output"`
	Tag       string   `short:"t" long:"tag" value-name:"job-name" description:"tag of the job"`
}

// Report is represents the result of the command
type Report struct {
	Command     string     `json:"command"`
	CommandArgs []string   `json:"commandArgs"`
	Tag         string     `json:"tag,omitempty"`
	Output      string     `json:"output"`
	Stdout      string     `json:"stdout"`
	Stderr      string     `json:"stderr"`
	ExitCode    *int       `json:"exitCode,omitempty"`
	Result      string     `json:"result"`
	Hostname    string     `json:"hostname"`
	Pid         *int       `json:"pid,omitempty"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
	SystemTime  *float64   `json:"systemTime,omitempty"`
	UserTime    *float64   `json:"userTime,omitempty"`
}

func (o *opts) run(args []string) (Report, error) {
	hostname, _ := os.Hostname()
	r := Report{
		Command:     shellquote.Join(args...),
		CommandArgs: args,
		Tag:         o.Tag,
		Hostname:    hostname,
	}
	cmd := exec.Command(args[0], args[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return o.failReport(r, err.Error()), err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		return o.failReport(r, err.Error()), err
	}

	var bufStdout bytes.Buffer
	var bufStderr bytes.Buffer
	var bufMerged bytes.Buffer

	var wtr io.Writer = &bufMerged
	if o.TimeStamp {
		wtr = newTimestampWriter(&bufMerged)
	}
	stdoutPipe2 := io.TeeReader(stdoutPipe, io.MultiWriter(&bufStdout, wtr))
	stderrPipe2 := io.TeeReader(stderrPipe, io.MultiWriter(&bufStderr, wtr))

	r.StartAt = now()
	err = cmd.Start()
	if err != nil {
		stderrPipe.Close()
		stdoutPipe.Close()
		return o.failReport(r, err.Error()), err
	}
	if cmd.Process != nil {
		r.Pid = &cmd.Process.Pid
	}
	done := make(chan struct{})
	go func() {
		o.runNoticer(r)
		done <- struct{}{}
	}()

	outDone := make(chan struct{})
	go func() {
		defer func() {
			stdoutPipe.Close()
			outDone <- struct{}{}
		}()
		io.Copy(os.Stdout, stdoutPipe2)
	}()

	errDone := make(chan struct{})
	go func() {
		defer func() {
			stderrPipe.Close()
			errDone <- struct{}{}
		}()
		io.Copy(os.Stderr, stderrPipe2)
	}()

	<-outDone
	<-errDone
	err = cmd.Wait()
	r.EndAt = now()
	ex := wrapcommander.ResolveExitCode(err)
	r.ExitCode = &ex
	r.Result = fmt.Sprintf("command exited with code: %d", *r.ExitCode)
	if *r.ExitCode > 128 {
		r.Result = fmt.Sprintf("command died with signal: %d", *r.ExitCode&127)
	}
	r.Stdout = bufStdout.String()
	r.Stderr = bufStderr.String()
	r.Output = bufMerged.String()
	if p := cmd.ProcessState; p != nil {
		durPtr := func(t time.Duration) *float64 {
			f := float64(t) / float64(time.Second)
			return &f
		}
		r.UserTime = durPtr(p.UserTime())
		r.SystemTime = durPtr(p.SystemTime())
	}
	o.runReporter(r)
	<-done

	return r, nil
}

func now() *time.Time {
	now := time.Now()
	return &now
}

func parseArgs(args []string) (*flags.Parser, *opts, []string, error) {
	o := &opts{}
	p := flags.NewParser(o, flags.Default)
	p.Usage = "--reporter /path/to/reporter.pl -- /path/to/job [...]\n\nVersion: " + version
	rest, err := p.ParseArgs(args)
	return p, o, rest, err
}

// Run the horenso
func Run(args []string) int {
	p, o, cmdArgs, err := parseArgs(args)
	if err != nil || len(cmdArgs) < 1 {
		if ferr, ok := err.(*flags.Error); !ok || ferr.Type != flags.ErrHelp {
			p.WriteHelp(os.Stderr)
		}
		return 2
	}
	r, err := o.run(cmdArgs)
	if err != nil {
		return wrapcommander.ResolveExitCode(err)
	}
	return *r.ExitCode
}

func (o *opts) failReport(r Report, errStr string) Report {
	fail := -1
	r.ExitCode = &fail
	r.Result = fmt.Sprintf("failed to execute command: %s", errStr)
	done := make(chan struct{})
	go func() {
		o.runNoticer(r)
		done <- struct{}{}
	}()
	o.runReporter(r)
	<-done
	return r
}

func runHandler(cmdStr string, json []byte) ([]byte, error) {
	args, err := shellquote.Split(cmdStr)
	if err != nil || len(args) < 1 {
		return nil, fmt.Errorf("invalid handler: %q", cmdStr)
	}
	cmd := exec.Command(args[0], args[1:]...)
	stdinPipe, _ := cmd.StdinPipe()
	stdinPipe.Write(json)
	stdinPipe.Close()
	return cmd.CombinedOutput()
}

func runHandlers(handlers []string, json []byte) {
	wg := &sync.WaitGroup{}
	for _, handler := range handlers {
		wg.Add(1)
		go func(h string) {
			runHandler(h, json)
			wg.Done()
		}(handler)
	}
	wg.Wait()
}

func (o *opts) runNoticer(r Report) {
	if len(o.Noticer) < 1 {
		return
	}
	json, _ := json.Marshal(r)
	runHandlers(o.Noticer, json)
}

func (o *opts) runReporter(r Report) {
	json, _ := json.Marshal(r)
	runHandlers(o.Reporter, json)
}
