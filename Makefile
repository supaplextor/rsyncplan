GOPATH=${HOME}/go:/usr/share/gocode
# Safer if not lowercase. Use CamelCase. (watch out in make clean)
APP=rsyncplan
# Safer if not empty. (watch out in make clean)
RELEASEDIR=release/

release:
	make build-one tgz GOOS=linux GOARCH=amd64
all: clean init release release-all
init:
	cd rsyncplan-go && go mod init ${APP}/v2 ; go mod tidy || true ; go mod download
	cd rsyncplan-exechook && go mod init ${APP}-exechook/v2 ; go mod tidy || true ; go mod download
build-one:
	cd rsyncplan-go && CGO_ENABLED=0 go build 
	cd rsyncplan-exechook && CGO_ENABLED=0 go build
release-all:
# part of 'release' --	make build-one tgz GOOS=linux GOARCH=amd64
#	make build-one tgz GOOS=android GOARCH=arm
#	make build-one tgz GOOS=darwin GOARCH=386	
	make build-one tgz GOOS=darwin GOARCH=amd64
#	make build-one tgz GOOS=darwin GOARCH=arm
	make build-one tgz GOOS=darwin GOARCH=arm64
#	make build-one tgz GOOS=plan9 GOARCH=386
#	make build-one tgz GOOS=plan9 GOARCH=amd64
	make build-one tgz GOOS=solaris GOARCH=amd64
	make build-one tgz GOOS=dragonfly GOARCH=amd64
	make build-one tgz GOOS=freebsd GOARCH=386
	make build-one tgz GOOS=freebsd GOARCH=amd64
	make build-one tgz GOOS=freebsd GOARCH=arm
	make build-one tgz GOOS=linux GOARCH=386
	make build-one tgz GOOS=linux GOARCH=arm
	make build-one tgz GOOS=linux GOARCH=arm64
	make build-one tgz GOOS=linux GOARCH=arm64
	make build-one tgz GOOS=linux GOARCH=ppc64
	make build-one tgz GOOS=linux GOARCH=ppc64le
	make build-one tgz GOOS=linux GOARCH=mips
	make build-one tgz GOOS=linux GOARCH=mipsle
	make build-one tgz GOOS=linux GOARCH=mips64
	make build-one tgz GOOS=linux GOARCH=mips64le
	make build-one tgz GOOS=netbsd GOARCH=386
	make build-one tgz GOOS=netbsd GOARCH=amd64
	make build-one tgz GOOS=netbsd GOARCH=arm
	make build-one tgz GOOS=openbsd GOARCH=386
	make build-one tgz GOOS=openbsd GOARCH=amd64
	make build-one tgz GOOS=openbsd GOARCH=arm
clean:
	cd rsyncplan-go && go clean
	cd rsyncplan-exechook && go clean
	rm -Rf ${RELEASEDIR} */go.sum */go.mod
bsd-install:
	cd rsyncplan-go && CGO_ENABLED=0 go build
	cd rsyncplan-exechook && CGO_ENABLED=0 go build
	install -v -d /usr/local/share/rsyncplan/
	install -v README.md /usr/local/share/rsyncplan/README.md
	install -v -p rsyncplan-go/rsyncplan rsyncplan-exechook/rsyncplan-exechook /usr/local/sbin/
gnu-install:
	cd rsyncplan-go && CGO_ENABLED=0 go build
	cd rsyncplan-exechook && CGO_ENABLED=0 go build
	install -v -d /usr/local/share/rsyncplan/
	install -m 0644 -v -p -t /usr/local/share/rsyncplan/ README.md
#	install -m 0644 -v -p -t /etc/sudoers.d/ rsyncplan-sudoers
	install -v -p -t /usr/local/sbin/ rsyncplan-go/rsyncplan rsyncplan-exechook/rsyncplan-exechook
uninstall:
	rm -v -f /usr/local/sbin/rsyncplan* /usr/local/share/rsyncplan/README.md /etc/sudoers.d/rsyncplan-sudoers
	rmdir -v /usr/local/share/rsyncplan/
tgz:
	test -d ${RELEASEDIR} || mkdir ${RELEASEDIR}
	cd .. && tar -zcvf ${APP}/${RELEASEDIR}/${APP}-release-${GOOS}-${GOARCH}.tgz ${APP}/README.md ${APP}/Makefile ${APP}/rsyncplan-go/rsyncplan ${APP}/rsyncplan-exechook/rsyncplan-exechook
