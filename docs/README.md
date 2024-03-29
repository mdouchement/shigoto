# 1. Shigoto

<!-- TOC -->

- [1. Shigoto](#1-shigoto)
    - [1.1. Overview](#11-overview)
    - [1.2. Runners](#12-runners)
        - [1.2.1. Exec](#121-exec)
        - [1.2.2. HTTP](#122-http)
        - [1.2.3. Shell](#123-shell)
        - [1.2.4. Yaegi](#124-yaegi)
        - [1.2.5. Defer](#125-defer)
        - [1.2.6. Tengo](#126-tengo)

<!-- /TOC -->


## 1.1. Overview

These files contain the scheduled tasks to process.
Each file is isolated so no variables or environment conflicts.

The Go’s template engine is used at several places in the Shigoto's YAML file.
All functions by the Go’s [sprig lib](http://masterminds.github.io/sprig/) are available.

```yaml
# Variables defines global templating variables used for all the tasks.
variables:
  MY_TEMPLATING_VAR: value1

# Environment defines global environment variables used for all the tasks.
environment:
  # Templating with the global variables.
  MY_ENVIRONMENT_VAR: "{{.MY_TEMPLATING_VAR}}"
  # Expanding with the host environment variables.
  WORKDIR: "${HOME}/workdir"

# The entrypoint for delaring your scheduled tasks.
shigoto:
  "task name":
    # Schedule defines when the task is run.
    # It follows standard crontab definition (`45 23 * * 6`).
    # It also support intervals (`@every <duration>` with duration a string accepted by
    #   [Go's duration parser](https://golang.org/pkg/time/#ParseDuration) like `1h30m10s`).
    schedule: "@every 5s"
    # Variables defines local templating variables used for the current task.
    # It supports templating using global templating variables as source.
    variables:
      MY_LOCAL_VAR: "{{.MY_TEMPLATING_VAR}}.log"
    # Environment defines local environment variables used for the current task.
    # It supports templating using global/local templating variables as source.
    # It supports envrironment expand using host/global envrironment variables as source.
    environment:
      WORKDIR: "{{.WORKDIR}}/${MY_ENVIRONMENT_VAR}"
      LOG_FILE: "/{{.MY_LOCAL_VAR}}"
    # Workdir is the FileSystem directory where the task works.
    # It supports global/local templating variables and host/global/local envrironment variables as source.
    workdir: /tmp
    # LogsFile - when defined - captures all stdout/stderr of the current task and redirect it to the given filename.
    # It supports global/local templating variables and host/global/local envrironment variables as source.
    # (default: stdout/stderr)
    logs_file: ${LOG_FILE}
    # Commands runs sequentially the given list of commands.
    # It supports global/local templating variables and host/global/local envrironment variables as source according the used runner.
    commands:
      - echo "shigoto ${LOG_FILE}"
```

## 1.2. Runners

### 1.2.1. Exec

[exec](https://golang.org/pkg/os/exec) runs binary command. It's the default runner used.

```yml
shigoto:
  baito_exec:
    schedule: "@every 5s"
    commands:
      # Without specifying the runner, `exec` will be used
      - echo "Hello Shigoto!"
      - exec: cat /tmp/no-such-file.log
        # Redirect the stdout/stderr of the current command to the given file.
        # (optional)
        redirect: /tmp/baito_1.log
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true
      - echo "The previous error has been ignored"
```

### 1.2.2. HTTP

[http](https://golang.org/pkg/net/http) runs HTTP request.

- No templating or environment expanding.

```yml
shigoto:
  baito_exec:
    schedule: "* * * * *"
    commands:
      - http: https://hc-ping.com/6ef113a2-UUID-UUID-UUID-{{.MY_VAR}}
        # Method HEAD, GET, POST, PUT, PATCH, DEETE are supported.
        # (default: get)
        method: get
        # Content-Type defines the content type of the body.
        # (optional)
        content_type: encoding/json
        # Body is the data to send to the URL.
        # (optional)
        body: |
          {
            "key": "value",
          }
        # Retry is the number of times the request is retried.
        # (default: 3)
        retry: 1
        # Interval is the duration between each retry.
        # It's a string accepted by [Go's duration parser](https://golang.org/pkg/time/#ParseDuration) like `1h30m10s`
        # (default: 20ms)
        retry_interval: 1s
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true
```

### 1.2.3. Shell

[Shell](https://github.com/mvdan/sh) allows to write Bash scripts.

- Supports global/local templating variables as source.
- Supports host/global/local envrironment variables as source.

```yml
shigoto:
  baito_shell:
    schedule: "@every 5s"
    variables:
      RANGE: |
        1
        2
        3
        4
        5
    commands:
      - sh: |
          for i in {{.RANGE | splitList "\n" | join " "}}
          do
            echo "Hello $i times"
          done
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true

```

### 1.2.4. Yaegi

[Yaegi](https://github.com/traefik/yaegi) allows to write Golang scripts.
Special import is `logger` that use the Shigoto [logger](https://godoc.org/github.com/mdouchement/logger#Logger).

- Supports global/local templating variables as source.
- Supports host/global/local envrironment variables as source.

```yml
shigoto:
  baito_yaegi:
    schedule: "@every 5s"
    commands:
      # Run a Golang file compatible with Yaegi.
      - yaegi: $HOME/main.go
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true
      # Declare Golang source code directly.
      - yaegi: |
          package main

          import (
            "logger"
          )

          func main() {
            logger.WithField("trololo", "popo").Info("Hello Yaegi")
          }
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true

```

### 1.2.5. Defer

With the `defer` keyword, it's possible to schedule cleanup to be run once the non deferred commands are completed.
The difference with just putting it as the last command is that this command will run even when the task fails.

```yml
shigoto:
  baito_defer:
    schedule: "@every 5s"
    commands:
      - defer: echo "defer 0"
      - defer: { exec: echo "defer 1" }
      - defer:
          sh: echo "defer 2"
      - echo "shigoto"

```

> Due to the nature of how the Go's own defer work, the deferred commands are executed in the reverse order if you schedule multiple of them.

### 1.2.6. Tengo

[Tengo](https://github.com/d5/tengo) allows to write Tengo scripts. It also uses extra libs/overwrites from [LDT](https://github.com/mdouchement/ldt).
Special import is `logger` that use the Shigoto [logger](https://godoc.org/github.com/mdouchement/logger#Logger).

- Supports global/local templating variables as source.
- Supports host/global/local envrironment variables as source.

```yml
variables:
  TENGO_LIST: |
    [
      "string 0",
      "string 1",
      "string 2"
    ]

shigoto:
  baito_tengo:
    schedule: "@every 5s"
    commands:
      # Run a Tengo file (*.tengo or *.tgo).
      - tengo: ~/shigoto.tengo
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true
      # Declare Tengo source code directly.
      - tengo: |
          fmt := import("fmt")

          list := {{.TENGO_LIST}}
          for v in list {
            fmt.println(v)
          }

          log := shigoto.logger
          log = log.with_prefix("[trololo]")
          log.info("done")
        # IgnoreError allows errors and continue to the next command.
        # (default: false)
        ignore_error: true

```