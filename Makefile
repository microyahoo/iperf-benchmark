# Normally, commands are expected in "cmd". That can be changed for a
# repository to something else by setting CMDS_DIR before including build.make.
CMDS_DIR ?= cmd

# The binary to build (just the basename).
BIN ?= iperf-benchmark


CMD := "go build -buildvcs=false -o ./bin/${BIN} ./${CMDS_DIR}/${BIN}"
build:  
	@go build -buildvcs=false -o ./bin/${BIN} ./${CMDS_DIR}/${BIN}

clean:
	@rm -rf ./bin

SKIP_TESTS ?=
test:
ifneq ($(SKIP_TESTS), 1)
	@go test -v `go list ./...`
endif
