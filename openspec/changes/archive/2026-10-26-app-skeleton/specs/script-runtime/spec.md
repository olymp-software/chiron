# Spec Delta

## MODIFIED Requirements

### Requirement: Guaranteed commands
On every OS a script SHALL be able to rely on bash builtins (`echo`, `printf`, `cd`, `test`/`[`, `export`, `exit`, etc.) and the utilities `cat cp mkdir mv rm touch`. These utilities SHALL be provided by the app and behave the same on Linux, macOS and Windows. They SHALL accept only the flags `mkdir -p`, `cp -r` and `rm -r -f`, combined short flags (`-rf`) and `--`; any other flag SHALL fail with a message on stderr and exit code 1. Anything else (`ls`, `find`, `chmod`, `xargs`, `grep`, `sed`, `awk`, `head`, `diff`, tools like `go`, `docker`, `kubectl`) is not guaranteed and SHALL be verified by the course's environment check.

#### Scenario: Guaranteed utility on Windows
- **WHEN** a script calls `cp a b` on Windows
- **THEN** the file is copied

#### Scenario: Same result on every OS
- **WHEN** a script runs `mkdir -p a/b && touch a/b/f && mkdir c && cp -r a c` on Linux, macOS and Windows
- **THEN** on each OS the file `c/a/b/f` exists

#### Scenario: Unsupported flag
- **WHEN** a script calls `cp -a a b`
- **THEN** the command prints a message to stderr, exits with code 1 and copies nothing

#### Scenario: Non-guaranteed utility on Windows
- **WHEN** a script calls `grep` on Windows without grep installed
- **THEN** the command fails as not found
