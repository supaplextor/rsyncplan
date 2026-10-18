# Security audit summary for rsyncplan

This audit reviews the code in `rsyncplan-go/rsyncplan.go` and `rsyncplan-exechook/rsyncplan-exechook.go` to identify the security-relevant measures that were examined while reviewing how the backup tool is constructed.

## Scope reviewed

- CLI input validation for required arguments and backup label/path values
- Construction of remote backup paths and use of `ssh`/`rsync`
- Execution of the helper binary via `--rsync-path`
- Directory creation and filesystem target selection
- Handling of rsync partial-transfer and verification failures
- Logging and operational safety around syslog/stderr output

## Measures investigated

### 1. Input validation and argument safety

The Go client enforces a few basic guardrails before starting the backup:

- It requires a destination host argument before proceeding.
- It rejects an empty `--label` value and rejects values containing `/`, so the backup label remains a single path component.
- It rejects an empty `--filesystem` value.
- It computes the remote backup base path from the local hostname and the validated label, rather than accepting arbitrary path text from the user.

This reduces the chances of path traversal or invalid remote snapshot naming caused by unchecked CLI input.

### 2. Limited remote path construction

The tool constructs the remote backup path as:

- `/backups/<hostname>/<label>/<timestamp>/`

This is not an arbitrary user-controlled path; it is built from the local host name and a sanitized label. The label is explicitly restricted to a single path component, which is the most important validation for keeping the remote backup hierarchy stable.

One caveat is that hostnames are accepted from the current system and are not independently normalized beyond the OS-provided value. In practice, the stronger trust boundary is the SSH configuration and the remote server's own path permissions.

### 3. Safe execution model around rsync

The program invokes `rsync` with an explicit helper binary:

- `--rsync-path=/usr/local/sbin/rsyncplan-exechook`

This design keeps the backup-side helper in a dedicated installed location instead of relying on an ad hoc shell command. The helper then runs `rsync` itself, preserving the remote `rsync` invocation path and allowing the program to handle exit codes in a controlled way.

The code also avoids `bash -c` style execution for the main backup step. Instead, it passes argument arrays directly to `exec.Command`, which reduces the risk of shell expansion or command-injection behavior from user-controlled strings.

### 4. Permission model for created directories

The helper creates the target directory with mode `0750`:

- `os.MkdirAll(lastArg, 0750)`

This is a restrictive directory permission that is appropriate for backup storage and reduces exposure of the created snapshot path to unrelated users on the local host.

### 5. Ignore rules and backup scope

The backup invocation includes:

- `--exclude=/swapfile`

This is a defensive measure to avoid copying the swap file, which is a common source of unnecessary or sensitive local data. The tool is intentionally scoped to a chosen filesystem root such as `/` or `/home`, not to an unrestricted recursive backup of all machine state.

### 6. rsync verification and partial-transfer handling

The code explicitly accounts for rsync behavior around file churn during a backup:

- exit code 23 and 24 are treated as non-fatal partial-transfer conditions
- the tool logs the condition and passes the exit status back to the caller so the client can report partial transfer status correctly

This is an important security and reliability measure because it avoids silently dropping data when files change while rsync is processing them. The code comments also note the `failed verification -- update discarded` condition, which is a known rsync behavior during concurrent filesystem mutation.

### 7. Logging and operational safety

The application configures logging to syslog with a fallback to stderr if syslog is unavailable:

- `syslog.New(...)`
- `log.SetOutput(logwriter)`

This keeps operational output in a standard system log path without requiring the tool to write debug or status data directly to the terminal. It also prevents a hard dependency on syslog availability from blocking the backup process, while still preserving some observability in failure situations.

The code does not appear to log credentials, SSH arguments, or other secret material in the reviewed code paths. The logger writes structured messages around path construction and rsync status without exposing token-like data.

### 8. Remote command execution boundary

The backup client runs a remote `ssh` command that lists recent snapshot directories:

- `ssh <host> "ls -1d <remote>/????-??-??*/ | sort -nr | head -n 20"`

This code is designed to enumerate existing backup snapshots, not to execute arbitrary user-controlled shell snippets. It still depends on the underlying SSH and remote shell being properly configured, and on the remote host being trusted. The specific safety property here is that the path prefix is generated from validated local inputs rather than arbitrary strings assembled from untrusted user data.

## Gaps and trust assumptions reviewed

The audit also considered the main limitations and assumptions in the current design:

- The code relies on the local SSH client and the remote SSH server to enforce authentication and host trust; it does not implement its own identity or host validation logic.
- The code applies basic validation to `--label` and `--filesystem`, but it does not perform full canonicalization or filesystem-escape checks for every possible input form.
- The backup tool intentionally assumes a trusted remote environment and a privileged backup account, so the security posture is only as strong as the SSH configuration and remote permissions.
- There is no chroot, sandbox, or privilege-drop step inside the Go process itself; this is a direct backup utility and not a hardened sandboxed daemon.

## Overall assessment

The reviewed code demonstrates a practical set of security-focused measures for a backup tool:

- input validation for key CLI parameters,
- constrained remote path generation,
- direct argument invocation instead of shell command construction,
- restrictive directory permissions,
- explicit handling of partial-transfer rsync failures,
- and logging to syslog with a safe fallback path.

These measures are consistent with a careful operational review of the backup workflow. The primary remaining security boundary is the trust placed in SSH and the remote backup server configuration, which is the normal operational model for this type of tool rather than a flaw in the local code alone.
