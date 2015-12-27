package horenso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
)

type opts struct {
	Reporter  []string `short:"r" long:"reporter" required:"true"`
	Noticer   []string `short:"n" long:"noticer"`
	TimeStamp bool     `short:"T" long:"timestamp"`
	Tag       string   `short:"t" long:"tag"`
}

type Report struct {
	Command     string     `json:"command"`
	CommandArgs []string   `json:"commandArgs"`
	Tag         string     `json:"tag,omitempty"`
	Output      string     `json:"output"`
	Stdout      string     `json:"stdout"`
	Stderr      string     `json:"stderr"`
	ExitCode    int        `json:"exitCode"`
	LineReport  string     `json:"lineReport"`
	Pid         int        `json:"pid"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
}

func (o *opts) run(args []string) (Report, error) {
	r := Report{
		Command:     shellquote.Join(args...),
		CommandArgs: args,
		Tag:         o.Tag,
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
	r.Pid = cmd.Process.Pid

	go func() {
		defer stdoutPipe.Close()
		io.Copy(os.Stdout, stdoutPipe2)
	}()

	go func() {
		defer stderrPipe.Close()
		io.Copy(os.Stderr, stderrPipe2)
	}()

	nCh := make(chan struct{})
	go func() {
		o.runNoticer(r)
		nCh <- struct{}{}
	}()

	err = cmd.Wait()
	r.EndAt = now()
	r.ExitCode = wrapcommander.ResolveExitCode(err)
	r.LineReport = fmt.Sprintf("command exited with code: %d", r.ExitCode)
	if r.ExitCode > 128 {
		r.LineReport = fmt.Sprintf("command died with signal: %d", r.ExitCode&127)
	}
	r.Stdout = bufStdout.String()
	r.Stderr = bufStderr.String()
	r.Output = bufMerged.String()
	o.runReporter(r)
	<-nCh

	return r, nil
}

func now() *time.Time {
	now := time.Now()
	return &now
}

func Run(args []string) int {
	optArgs, cmdArgs := wrapcommander.SeparateArgs(args)
	o, err := parseArgs(optArgs)
	if err != nil {
		return 2
	}
	r, err := o.run(cmdArgs)
	if err != nil {
		return wrapcommander.ResolveExitCode(err)
	}
	return r.ExitCode
}

func (o *opts) failReport(r Report, errStr string) Report {
	r.ExitCode = -1
	r.LineReport = fmt.Sprintf("failed to execute command: %s", errStr)
	o.runNoticer(r)
	o.runReporter(r)
	return r
}

func runHandler(cmdStr string, r Report) ([]byte, error) {
	args, err := shellquote.Split(cmdStr)
	if err != nil || len(args) < 1 {
		log.Printf("invalid handler %q", cmdStr)
		return nil, nil
	}
	byt, _ := json.Marshal(r)
	prog := args[0]
	argv := append(args[1:], string(byt))
	cmd := exec.Command(prog, argv...)
	return cmd.CombinedOutput()
}

func runHandlers(handlers []string, r Report) {
	wg := &sync.WaitGroup{}
	for _, handler := range handlers {
		wg.Add(1)
		go func(h string) {
			runHandler(h, r)
			wg.Done()
		}(handler)
	}
	wg.Wait()
}

func (o *opts) runNoticer(r Report) {
	if len(o.Noticer) < 1 {
		return
	}
	runHandlers(o.Noticer, r)
}

func (o *opts) runReporter(r Report) {
	runHandlers(o.Reporter, r)
}

func parseArgs(args []string) (*opts, error) {
	opts := &opts{}
	_, err := flags.ParseArgs(opts, args)
	return opts, err
}
