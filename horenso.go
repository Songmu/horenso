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
	"strings"
	"time"

	"github.com/Songmu/timestamper"
	"github.com/Songmu/wrapcommander"
	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
	"github.com/lestrrat-go/strftime"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/transform"
)

type horenso struct {
	Reporter       []string `short:"r" long:"reporter" value-name:"/path/to/reporter.pl" description:"handler for reporting the result of the job"`
	Noticer        []string `short:"n" long:"noticer" value-name:"'ruby /path/to/noticer.rb'" description:"handler for noticing the start of the job"`
	TimeStamp      bool     `short:"T" long:"timestamp" description:"add timestamp to merged output"`
	Tag            string   `short:"t" long:"tag" value-name:"job-name" description:"tag of the job"`
	OverrideStatus bool     `short:"o" long:"override-status" description:"override command exit status, always exit 0"`
	Verbose        []bool   `short:"v" long:"verbose" description:"verbose output. it can be stacked like -vv for more detailed log"`
	Logfile        string   `short:"l" long:"log" value-name:"/path/to/logfile" description:"logfile path. The strftime format like '%Y%m%d.log' is available."`
	Config         string   `short:"c" long:"config" value-name:"/path/to/config.yaml" description:"config file"`

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
	ExitCode    int        `json:"exitCode"`
	Signaled    bool       `json:"signaled"`
	Result      string     `json:"result"`
	Hostname    string     `json:"hostname"`
	Pid         int        `json:"pid,omitempty"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
	SystemTime  float64    `json:"systemTime,omitempty"`
	UserTime    float64    `json:"userTime,omitempty"`
}

func (ho *horenso) openLog() (io.WriteCloser, error) {
	logfile, err := strftime.Format(ho.Logfile, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to parse log file format %q: %s", ho.Logfile, err)
	}
	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %s", logfile, err)
	}
	return f, nil
}

func (ho *horenso) loadConfig() error {
	conf := ho.Config
	if conf == "" {
		conf = os.Getenv("HORENSO_CONFIG")
	}
	if conf == "" {
		return nil
	}
	c, err := loadConfig(conf)
	if err != nil {
		return err
	}
	ho.Reporter = append(ho.Reporter, c.Reporter...)
	ho.Noticer = append(ho.Noticer, c.Noticer...)
	if !ho.TimeStamp {
		ho.TimeStamp = c.Timestamp
	}
	if ho.Tag == "" {
		ho.Tag = c.Tag
	}
	if !ho.OverrideStatus {
		ho.OverrideStatus = c.OverrideStatus
	}
	if ho.Logfile == "" {
		ho.Logfile = c.Logfile
	}
	return nil
}

func (ho *horenso) run(args []string) (Report, error) {
	log.SetPrefix("[horenso] ")
	log.SetFlags(0)
	log.SetOutput(ho.errStream)

	if err := ho.loadConfig(); err != nil {
		ho.logf(warn, "failed to load config: %s", err)
	}

	hostname, _ := os.Hostname()
	r := Report{
		Command:     shellquote.Join(args...),
		CommandArgs: args,
		Tag:         ho.Tag,
		ExitCode:    -1,
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
	if ho.Logfile != "" {
		if f, err := ho.openLog(); err != nil {
			ho.log(warn, err.Error())
		} else {
			defer f.Close()
			wtr = io.MultiWriter(wtr, f)
		}
	}
	if ho.TimeStamp {
		wc := transform.NewWriter(wtr, timestamper.New())
		defer wc.Close()
		wtr = wc
	}
	stdoutPipe2 := io.TeeReader(stdoutPipe, io.MultiWriter(&bufStdout, wtr))
	stderrPipe2 := io.TeeReader(stderrPipe, io.MultiWriter(&bufStderr, wtr))

	ho.logf(info, "starting execution of the command %q", r.Command)
	r.StartAt = now()
	err = cmd.Start()
	if err != nil {
		return ho.failReport(r, err.Error()), err
	}
	if cmd.Process != nil {
		r.Pid = cmd.Process.Pid
	}
	done := make(chan error)
	go func(r Report) {
		done <- ho.runNoticer(r)
	}(r)

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
	if err := eg.Wait(); err != nil {
		ho.logf(warn, "something went wrong while executing the command: %s", err)
	}
	err = cmd.Wait()
	r.EndAt = now()
	es := wrapcommander.ResolveExitStatus(err)
	r.ExitCode = es.ExitCode()
	r.Signaled = es.Signaled()
	r.Result = fmt.Sprintf("command exited with code: %d", r.ExitCode)
	if r.Signaled {
		r.Result = fmt.Sprintf("command died with signal: %d", r.ExitCode&127)
	}
	ho.logf(info, "the command %q finished: %s", r.Command, r.Result)
	r.Stdout = bufStdout.String()
	r.Stderr = bufStderr.String()
	r.Output = bufMerged.String()
	if p := cmd.ProcessState; p != nil {
		r.UserTime = float64(p.UserTime()) / float64(time.Second)
		r.SystemTime = float64(p.SystemTime()) / float64(time.Second)
	}
	ho.runReporter(r)
	<-done
	ho.logf(info, "all processes are completed for the job %q", r.Command)
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
	return r.ExitCode
}

func (ho *horenso) failReport(r Report, errStr string) Report {
	r.Result = fmt.Sprintf("failed to execute the command: %s", errStr)
	ho.logf(warn, "failed to execute the command %q: %s", r.Command, errStr)
	done := make(chan error)
	go func() {
		done <- ho.runNoticer(r)
	}()
	ho.runReporter(r)
	<-done
	return r
}

func (ho *horenso) appendOut(base, out string) string {
	out = strings.TrimSpace(out)
	if out == "" {
		return base
	}
	if !strings.HasSuffix(base, "\n") {
		base += "\n"
	}
	const indent = "  "
	return base + indent + strings.Replace("Output:\n"+out, "\n", "\n"+indent, -1)
}

func (ho *horenso) splitHandlerCmdStr(cmdStr string) ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		args := strings.Split(cmdStr, " ")
		return args, nil
	default:
		args, err := shellquote.Split(cmdStr)
		return args, err
	}
}

func (ho *horenso) runHandler(cmdStr string, json []byte) error {
	ho.logf(info, "starting to run the handler %q", cmdStr)
	args, err := ho.splitHandlerCmdStr(cmdStr)
	if err != nil || len(args) < 1 {
		ho.logf(warn, "failed to run the handler %q: invalid handler arguments", cmdStr)
		return fmt.Errorf("invalid handler: %q", cmdStr)
	}
	cmd := exec.Command(args[0], args[1:]...)
	stdinPipe, _ := cmd.StdinPipe()
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Start(); err != nil {
		stdinPipe.Close()
		logoutput := fmt.Sprintf("failed to run the handler %q: %s", cmdStr, err)
		ho.log(warn, ho.appendOut(logoutput, b.String()))
		return err
	}
	stdinPipe.Write(json)
	stdinPipe.Close()
	err = cmd.Wait()
	if err != nil || ho.logLevel() >= info {
		var logoutput string
		lv := info
		if err != nil {
			lv = warn
			logoutput = fmt.Sprintf("failed to run the handler %q: %s", cmdStr, err)
		} else {
			logoutput = fmt.Sprintf("finished to run the handler %q", cmdStr)
		}
		ho.log(lv, ho.appendOut(logoutput, b.String()))
	}
	return err
}

func (ho *horenso) runHandlers(handlers []string, json []byte) error {
	eg := &errgroup.Group{}
	for _, handler := range handlers {
		h := handler
		eg.Go(func() error {
			return ho.runHandler(h, json)
		})
	}
	return eg.Wait()
}

func (ho *horenso) runNoticer(r Report) error {
	if len(ho.Noticer) < 1 {
		return nil
	}
	ho.logf(info, "starting to run the noticers")
	defer ho.logf(info, "finished to run the noticers")
	json, _ := json.Marshal(r)
	return ho.runHandlers(ho.Noticer, json)
}

func (ho *horenso) runReporter(r Report) error {
	ho.logf(info, "starting to run the reporters")
	defer ho.logf(info, "finished to run the reporters")
	json, _ := json.Marshal(r)
	return ho.runHandlers(ho.Reporter, json)
}
