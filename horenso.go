package horenso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
	"golang.org/x/sync/errgroup"
)

type horenso struct {
	Reporter       []string `short:"r" long:"reporter" required:"true" value-name:"/path/to/reporter.pl" description:"handler for reporting the result of the job"`
	Noticer        []string `short:"n" long:"noticer" value-name:"/path/to/noticer.rb" description:"handler for noticing the start of the job"`
	TimeStamp      bool     `short:"T" long:"timestamp" description:"add timestamp to merged output"`
	Tag            string   `short:"t" long:"tag" value-name:"job-name" description:"tag of the job"`
	OverrideStatus bool     `short:"o" long:"override-status" description:"override command exit status, always exit 0"`
	Verbose        []bool   `short:"v" long:"verbose" description:"verbose output"`

	outStream, errStream io.Writer
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
	Signaled    bool       `json:"signaled"`
	Result      string     `json:"result"`
	Hostname    string     `json:"hostname"`
	Pid         *int       `json:"pid,omitempty"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
	SystemTime  *float64   `json:"systemTime,omitempty"`
	UserTime    *float64   `json:"userTime,omitempty"`
}

func (ho *horenso) run(args []string) (Report, error) {
	log.SetOutput(ho.errStream)

	hostname, _ := os.Hostname()
	r := Report{
		Command:     shellquote.Join(args...),
		CommandArgs: args,
		Tag:         ho.Tag,
		Hostname:    hostname,
	}
	cmd := exec.Command(args[0], args[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return ho.failReport(r, err.Error()), err
	}
	defer stdoutPipe.Close()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return ho.failReport(r, err.Error()), err
	}
	defer stderrPipe.Close()

	var bufStdout bytes.Buffer
	var bufStderr bytes.Buffer
	var bufMerged bytes.Buffer

	var wtr io.Writer = &bufMerged
	if ho.TimeStamp {
		wtr = newTimestampWriter(&bufMerged)
	}
	stdoutPipe2 := io.TeeReader(stdoutPipe, io.MultiWriter(&bufStdout, wtr))
	stderrPipe2 := io.TeeReader(stderrPipe, io.MultiWriter(&bufStderr, wtr))

	r.StartAt = now()
	err = cmd.Start()
	if err != nil {
		return ho.failReport(r, err.Error()), err
	}
	if cmd.Process != nil {
		r.Pid = &cmd.Process.Pid
	}
	done := make(chan error)
	go func() {
		done <- ho.runNoticer(r)
	}()

	eg := &errgroup.Group{}
	eg.Go(func() error {
		defer stdoutPipe.Close()
		_, err := io.Copy(ho.outStream, stdoutPipe2)
		return err
	})
	eg.Go(func() error {
		defer stderrPipe.Close()
		_, err := io.Copy(ho.errStream, stderrPipe2)
		return err
	})
	eg.Wait()

	err = cmd.Wait()
	r.EndAt = now()
	es := wrapcommander.ResolveExitStatus(err)
	ecode := es.ExitCode()
	r.ExitCode = &ecode
	r.Signaled = es.Signaled()
	r.Result = fmt.Sprintf("command exited with code: %d", *r.ExitCode)
	if r.Signaled {
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
	ho.runReporter(r)
	<-done

	return r, nil
}

func now() *time.Time {
	now := time.Now()
	return &now
}

func parseArgs(args []string) (*flags.Parser, *horenso, []string, error) {
	ho := &horenso{}
	p := flags.NewParser(ho, flags.Default)
	p.Usage = fmt.Sprintf(`--reporter /path/to/reporter.pl -- /path/to/job [...]

Version: %s (rev: %s/%s)`, version, revision, runtime.Version())
	rest, err := p.ParseArgs(args)
	ho.outStream = os.Stdout
	ho.errStream = os.Stderr
	return p, ho, rest, err
}

// Run the horenso
func Run(args []string) int {
	log.SetPrefix("[horenso] ")
	log.SetFlags(0)

	p, ho, cmdArgs, err := parseArgs(args)
	if err != nil || len(cmdArgs) < 1 {
		if ferr, ok := err.(*flags.Error); !ok || ferr.Type != flags.ErrHelp {
			p.WriteHelp(ho.errStream)
		}
		return 2
	}
	r, err := ho.run(cmdArgs)
	if err != nil {
		return wrapcommander.ResolveExitCode(err)
	}
	if ho.OverrideStatus {
		return 0
	}
	return *r.ExitCode
}

func (ho *horenso) failReport(r Report, errStr string) Report {
	fail := -1
	r.ExitCode = &fail
	r.Result = fmt.Sprintf("failed to execute command: %s", errStr)
	done := make(chan error)
	go func() {
		done <- ho.runNoticer(r)
	}()
	ho.runReporter(r)
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
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Start(); err != nil {
		stdinPipe.Close()
		return b.Bytes(), err
	}
	stdinPipe.Write(json)
	stdinPipe.Close()
	err = cmd.Wait()
	return b.Bytes(), err
}

func (ho *horenso) runHandlers(handlers []string, json []byte) error {
	eg := &errgroup.Group{}
	for _, handler := range handlers {
		h := handler
		eg.Go(func() error {
			_, err := runHandler(h, json)
			return err
		})
	}
	return eg.Wait()
}

func (ho *horenso) runNoticer(r Report) error {
	if len(ho.Noticer) < 1 {
		return nil
	}
	json, _ := json.Marshal(r)
	return ho.runHandlers(ho.Noticer, json)
}

func (ho *horenso) runReporter(r Report) error {
	json, _ := json.Marshal(r)
	return ho.runHandlers(ho.Reporter, json)
}
