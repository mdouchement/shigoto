# Shigoto

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/shigoto)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/shigoto)](https://goreportcard.com/report/github.com/mdouchement/shigoto)
[![License](https://img.shields.io/github/license/mdouchement/shigoto.svg)](http://opensource.org/licenses/MIT)

A nextgen crontab.
It's a Golang alternative of crontab heavily inspired of [Taskfile](https://taskfile.dev) templating mechanism.


### Technologies / Frameworks

- [Cobra](https://github.com/spf13/cobra)
- [Koanf](https://github.com/knadh/koanf)
- [Cron](https://github.com/robfig/cron)
- [sh](https://github.com/mvdan/sh)
- [Yaegi](https://github.com/containous/yaegi)

## Shigoto file

File must be in the folder defined by the directive `directory` from Shigoto configuration file.

```yml
variables:
  MY_VAR: value1

environment:
  MY_VAR: "{{.MY_VAR}}"

shigoto:
  "Do something":
    schedule: "@every 10s"
    workdir: /tmp
    logs_file: /tmp/trololo.log
    variables:
      MY_VAR: value2
    commands:
      - touch {{.MY_VAR}}.txt
```

More [documentation](https://github.com/mdouchement/shigoto/tree/master/docs).

## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
