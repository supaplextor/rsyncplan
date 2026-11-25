When crashplan is too expensive, you rewrite shell scripts

# Installing

## apt based installs
* <code>sudo apt install git make rsync ssh rsyslog</code>
* <code>git clone https://github.com/supaplextor/rsyncplan.git</code>
* <code>cd rsyncplan</code>
* <code>make clean init release</code>
* <code>sudo make gnu-install</code> or <code># make bsd-install</code>

Note: `make release-all` builds cross-compiled release tarballs for all
supported platforms and is **not** needed for a single-machine install.
The install targets (`gnu-install`, `bsd-install`) always rebuild the
binaries natively for the current host architecture before installing,
so you do not need to run `release-all` first.

## MacOS
## FreeBSD installs
## Fedora
## OpenSUSE

# Sample Invocation
<code># rsyncplan rsyncplan-plus201.rollback.cloud</code>

<code># rsyncplan --label home --filesystem=/home/ rsyncplan-plus201.rollback.cloud</code>

Defaults:
* <code>--label=rootfs</code>
* <code>--filesystem=/</code>

It will fire off one ssh connection and close to grab most recent
directory list. Next it will call rsync. You're advised to setup
a ssh config entry for the remote root shell. This allows
passwordless login via ssh keys.

# rsync --link-dest= strategy

We collect a ls -1d and parse present directories 
as --link-dest= args

# Target server directory structure
* /backups
* /backups/machine-name/label/YYYY-MM-DD_HHMMSS_NS

Server side (say the above with sshd, rsync) can use 
hardlinks with recent <code>yyyy-mm-dd</code> attempts 
to keep disk space use lower between different snapshots.

# rsyncplan compared with Timeshift

[Timeshift](https://github.com/linuxmint/timeshift) is a Linux system
restore tool. It can create scheduled snapshots using rsync and hard links,
or Btrfs snapshots on supported Btrfs layouts, and provides workflows for
restoring a system from those snapshots.

rsyncplan is a command-line backup transfer tool. It runs rsync over
SSH to a separate backup server, creating timestamped directories and using
recent snapshots as `--link-dest` targets to reuse unchanged files. You choose
when to run it and which filesystem or path to send; it does not schedule
snapshots or provide a system restore interface. It does not apply Timeshift's
default exclusions for user home directories; files within the selected
filesystem are in scope unless excluded by rsync options.

In short, choose Timeshift for convenient Linux system snapshots and restore.
Choose rsyncplan when you want manually run, SSH-based copies
to a separately managed backup server. They solve different problems and can
be used together.

# "failed verification -- update discarded" errors

rsync performs an end-to-end checksum verification after transferring each file.
If a source file changes *while* rsync is reading and sending it (for example,
a compiler or build system regenerates files in an <code>*_autogen/</code>
directory concurrently with the backup), the checksums will not match and rsync
reports:

<pre>ERROR: path/to/file failed verification -- update discarded.</pre>

rsync then exits with code **23** (partial transfer due to error) or **24**
(partial transfer due to vanished source files).  These are non-fatal: the
affected files will simply be backed up correctly on the next rsync run once
they are no longer changing.  rsyncplan treats exit codes 23 and 24 as
warnings and continues normally.
