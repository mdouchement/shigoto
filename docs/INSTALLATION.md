# Installation

## Binary

- Download from https://github.com/mdouchement/shigoto/releases/latest
  - Or compile via Taskfile or `go build cmd/shigoto/main.go`
- Put the binary in `/usr/sbin/shigoto`.

## Configuration

This file contains the configuration of Shigoto.
The default directories lookup are:
- `/etc/shigoto/shigoto.toml`
- `/shigoto/shigoto.toml`
- `./shigoto.toml`

```toml
# The directory where the Shigoto's YAML files are.
directory = "/etc/shigoto"
# The socket mainly used for relaoding Shigoto's daemon.
socket = "/var/run/shigoto.sock"

[log]
# Force the colo in non-tty caller
force_color = true
# Force the colo in non-tty caller
force_formating = true
```

## Systemd

`/lib/systemd/system/shigoto.service`

```toml
[Unit]
Description=Shigoto, a nextgen crontab
After=network.target

[Service]
PIDFile=/run/shigoto.pid
ExecStart=/usr/sbin/shigoto daemon
ExecReload=/usr/sbin/shigoto reload
ExecStop=/bin/kill -s TERM $MAINPID

[Install]
WantedBy=multi-user.target
```

> Logs: `journalctl --unit shigoto`