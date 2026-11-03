horenso(報・連・相)
===================

[![Test Status](https://github.com/Songmu/horenso/workflows/test/badge.svg?branch=main)][actions]
[![codecov.io](https://codecov.io/github/Songmu/horenso/coverage.svg?branch=main)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/horenso)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/horenso)][PkgGoDev]

[actions]: https://github.com/Songmu/horenso/actions?workflow=test
[codecov]: https://codecov.io/github/Songmu/horenso?branch=main
[license]: https://github.com/Songmu/horenso/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/horenso

## Description

Command wrapper for reporting the result. It is useful for cron jobs.

## Installation

    % go get github.com/Songmu/horenso/cmd/horenso

Built binaries are available on gihub releases.
<https://github.com/Songmu/horenso/releases>

You can also install horenso with [aqua](https://aquaproj.github.io/).

    % aqua g -i Songmu/horenso

## Synopsis

    % horenso --reporter /path/to/report.pl -- /path/to/yourjob

## Options

```
Usage:
  horenso --reporter /path/to/reporter.pl -- /path/to/job [...]

Application Options:
  -r, --reporter=/path/to/reporter.pl      handler for reporting the result of the job
  -n, --noticer='ruby /path/to/noticer.rb' handler for noticing the start of the job
  -T, --timestamp                          add timestamp to merged output
  -t, --tag=job-name                       tag of the job
  -o, --override-status                    override command exit status, always exit 0
  -v, --verbose                            verbose output. it can be stacked like -vv for
                                           more detailed log
  -l, --log=logfile-path                   logfile path. The strftime format like
                                           '%Y%m%d.log' is available.
```

Handlers are should be an executable or command line string. You can specify multiple reporters and noticers.
In this case, they are executed concurrently.

## Usage

Normally you can use `horenso` with a wrapper shell script like following.

```shell
#!/bin/bash
/path/to/horenso \
  -n /path/to/noticer.py         \
  -r /path/to/reporter.pl        \
  -r 'ruby /path/to/reporter.rb' \
  -- "$@"
```

And specify this `wrapper.sh` in the crontab like following.

```
3 4 * * * /path/to/wrapper.sh /path/to/job --args...
```

If you want to change reporting way, you just have to change reporter script. You have no risk to crash
wrapper shell.

## Execution Sequence

1. Start the command
2. [optional] Run the noticers
3. Wait to finish the command
4. Run the reporters

## result JSON

The reporters and noticers accept a result JSON via STDIN that reports command result like following.

```json
{
  "command": "perl -E 'say 1;warn \"$$\\n\";'",
  "commandArgs": [
    "perl",
    "-E",
    "say 1;warn \"$$\\n\";"
  ],
  "output": "1\n95030\n",
  "stdout": "1\n",
  "stderr": "95030\n",
  "exitCode": 0,
  "result": "command exited with code: 0",
  "pid": 95030,
  "startAt": "2015-12-28T00:37:10.494282399+09:00",
  "endAt": "2015-12-28T00:37:10.546466379+09:00",
  "hostname": "webserver.example.com",
  "systemTime": 0.034632,
  "userTime": 0.026523
}
```

## License

[MIT][license]

## Author

[Songmu](https://github.com/Songmu)
