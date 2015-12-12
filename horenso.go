package horenso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
)

type opts struct {
	Reporter string `long:"reporter" required:"true"`
}

type Report struct {
	Output     string `json:"output"`
	Command    string `json:"command"`
	ExitCode   *int   `json:"exitCode,omitempty"`
	LineReport string `json:"lineReport"`
}

func Run(args []string) int {
	optArgs, cmdArgs := wrapcommander.SeparateArgs(args)
	o, err := parseArgs(optArgs)
	if err != nil {
		return 2
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

	err = cmd.Start()
	if err != nil {
		stderrPipe.Close()
		stdoutPipe.Close()
		o.failReport(cmdArgs, err.Error())
		return wrapcommander.ResolveExitCode(err)
	}

	go func() {
		defer stdoutPipe.Close()
		io.Copy(os.Stdout, stdoutPipe2)
	}()

	go func() {
		defer stderrPipe.Close()
		io.Copy(os.Stderr, stderrPipe2)
	}()

	err = cmd.Wait()
	fmt.Printf("%+v\n", err)
	exitCode := wrapcommander.ResolveExitCode(err)

	lineReport := fmt.Sprintf("command exited with code: %d", exitCode)
	if exitCode > 128 {
		lineReport = fmt.Sprintf("command died with signal: %d", exitCode&127)
	}
	report := Report{
		Output:     bufMerged.String(),
		Command:    shellquote.Join(cmdArgs...),
		ExitCode:   &exitCode,
		LineReport: lineReport,
	}
	o.runReporter(report)
	return exitCode
}

func (o *opts) failReport(cmdArgs []string, errStr string) {
	report := Report{
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
