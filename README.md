When crashplan is too expensive, you rewrite shell scripts

# Installing

## apt based installs
* <code>sudo apt install git make rsync ssh rsyslog</code>
* <code>git clone https://github.com/supaplextor/rsyncplan.git</code>
* <code>cd rsyncplan</code>
* <code>make clean init release</code> or <code>make clean init release-all</code>
* <code>sudo make gnu-install</code> or <code># make bsd-install</code>

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
