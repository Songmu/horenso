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
	Reporter string `long:"reporter" required:"true"`
	Noticer  string `long:"noticer"`
}

type Report struct {
	Command    string    `json:"command"`
	Output     string    `json:"output"`
	ExitCode   int       `json:"exitCode"`
	LineReport string    `json:"lineReport"`
	Pid        int       `json:"pid"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
}

func Run(args []string) int {
	optArgs, cmdArgs := wrapcommander.SeparateArgs(args)
	o, err := parseArgs(optArgs)
	if err != nil {
		return 2
	}

	r := Report{
		Command: shellquote.Join(cmdArgs...),
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		o.failReport(cmdArgs, err.Error())
		return wrapcommander.ResolveExitCode(err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		o.failReport(cmdArgs, err.Error())
		return wrapcommander.ResolveExitCode(err)
	}

	var bufMerged bytes.Buffer
	stdoutPipe2 := io.TeeReader(stdoutPipe, &bufMerged)
	stderrPipe2 := io.TeeReader(stderrPipe, &bufMerged)

	r.StartAt = time.Now()
	err = cmd.Start()
	if err != nil {
		stderrPipe.Close()
		stdoutPipe.Close()
		o.failReport(cmdArgs, err.Error())
		return wrapcommander.ResolveExitCode(err)
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

	err = cmd.Wait()
	r.EndAt = time.Now()
	r.ExitCode = wrapcommander.ResolveExitCode(err)
	r.LineReport = fmt.Sprintf("command exited with code: %d", r.ExitCode)
	if r.ExitCode > 128 {
		r.LineReport = fmt.Sprintf("command died with signal: %d", r.ExitCode&127)
	}
	r.Output = bufMerged.String()
	o.runReporter(r)
	return r.ExitCode
}

func (o *opts) failReport(cmdArgs []string, errStr string) {
	report := Report{
		ExitCode:   -1,
		Command:    shellquote.Join(cmdArgs...),
		LineReport: fmt.Sprintf("failed to execute command: %s", errStr),
	}
	o.runReporter(report)
}

func (o *opts) runReporter(report Report) {
	args, err := shellquote.Split(o.Reporter)
	if err != nil {
		log.Print(err)
		return
	}
	byt, _ := json.Marshal(report)
	if len(args) < 1 {
		log.Println("no reporter specified")
		return
	}
	prog := args[0]
	argv := append(args[1:], string(byt))
	cmd := exec.Command(prog, argv...)
	out, err := cmd.CombinedOutput()
	// DEBUG
	log.Println(string(out))
	if err != nil {
		log.Print(err)
	}
}

func parseArgs(args []string) (*opts, error) {
	opts := &opts{}
	_, err := flags.ParseArgs(opts, args)
	return opts, err
}
