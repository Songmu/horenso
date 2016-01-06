horenso(報・連・相)
===================

[![Build Status](https://travis-ci.org/Songmu/horenso.png?branch=master)][travis]
[![Coverage Status](https://coveralls.io/repos/Songmu/horenso/badge.png?branch=master)][coveralls]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[travis]: https://travis-ci.org/Songmu/horenso
[coveralls]: https://coveralls.io/r/Songmu/horenso?branch=master
[license]: https://github.com/Songmu/horenso/blob/master/LICENSE

## Description

Command wrapper for reporting the result. It is useful for cron jobs.

**THE SOFTWARE IS IT'S IN ALPHA QUALITY. IT MAY CHANGE THE API WITHOUT NOTICE.**

## Installation

    % go get github.com/Songmu/horenso/cmd/horenso

Built binaries are available on gihub releases soon!

## Synopsis

    % horenso --reporter /path/to/report.pl -- /path/to/yourjob

## Options

```
Usage:
  horenso --reporter /path/to/reporter.pl -- /path/to/job [...]

Application Options:
  -r, --reporter=/path/to/reporter.pl     handler for reporting the result of the job
  -n, --noticer='ruby/path/to/noticer.rb' handler for noticing the start of the job
  -T, --timestamp                         add timestamp to merged output
  -t, --tag=job-name                      tag of the job
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

## Execution Seqence

1. Start the command
2. [optional] Run the noticers
3. Wait to finish the command
4. Run the reporters

## JSON Argument

The reporters and noticers are accept one argument of JSON which reports command result like following.

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
  "hostname": "webserver.mydomain.com"
}
```

## License

[MIT][license]

## Author

[Songmu](https://github.com/Songmu)
