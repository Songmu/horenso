package horenso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
)

type opts struct {
	Reporter  string `long:"reporter" required:"true"`
	Noticer   string `long:"noticer"`
	TimeStamp bool   `long:"timestamp"`
}

type Report struct {
	Command    string    `json:"command"`
	Output     string    `json:"output"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	ExitCode   int       `json:"exitCode"`
	LineReport string    `json:"lineReport"`
	Pid        int       `json:"pid"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
}

func (o *opts) run(args []string) Report {
	r := Report{
		Command: shellquote.Join(args...),
	}
	cmd := exec.Command(args[0], args[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return o.failReport(r, err.Error())
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		return o.failReport(r, err.Error())
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

	r.StartAt = time.Now()
	err = cmd.Start()
	if err != nil {
		stderrPipe.Close()
		stdoutPipe.Close()
		return o.failReport(r, err.Error())
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
	o.runNoticer(r)

	err = cmd.Wait()
	r.EndAt = time.Now()
	r.ExitCode = wrapcommander.ResolveExitCode(err)
	r.LineReport = fmt.Sprintf("command exited with code: %d", r.ExitCode)
	if r.ExitCode > 128 {
		r.LineReport = fmt.Sprintf("command died with signal: %d", r.ExitCode&127)
	}
	r.Stdout = bufStdout.String()
	r.Stderr = bufStderr.String()
	r.Output = bufMerged.String()
	o.runReporter(r)

	return r
}

func Run(args []string) int {
	optArgs, cmdArgs := wrapcommander.SeparateArgs(args)
	o, err := parseArgs(optArgs)
	if err != nil {
		return 2
	}
	r := o.run(cmdArgs)
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

func (o *opts) runNoticer(r Report) {
	if o.Noticer == "" {
		return
	}
	out, _ := runHandler(o.Noticer, r)
	// DEBUG
	fmt.Println(string(out))
}

func (o *opts) runReporter(r Report) {
	out, _ := runHandler(o.Reporter, r)
	// DEBUG
	fmt.Println(string(out))
}

func parseArgs(args []string) (*opts, error) {
	opts := &opts{}
	_, err := flags.ParseArgs(opts, args)
	return opts, err
}
