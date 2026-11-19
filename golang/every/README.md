# every

Run a command only if enough time has passed since its last run.

```sh
every day : git pull            # once per calendar day
every 120s : date               # at least 120 seconds apart
every -t backup week : rsync …  # tracked under the tag "backup"
every                           # list the tracked commands
every -h                        # a terse help
every -H                        # the full help
```

A Go port of the Bash script [`shell/every`](../../shell/every), with the same options, messages and exit status. The two share the same db, `${XDG_STATE_HOME:-~/.local/state}/every/last-run.tsv`, and the same lock, so they can be used side by side. Commands are keyed in the db by their quoted text as `printf %q` of `/bin/bash` 3.2 (macOS) writes it.

```sh
go install github.com/squillero/programmer-zendo/golang/every@latest
go test .
```
