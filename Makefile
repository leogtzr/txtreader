.DEFAULT_GOAL := install

# INSTALL_SCRIPT=./install.sh
BIN_FILE=txtreader

install:
	go build -o "${BIN_FILE}"

clean:
	go clean
	rm -f "cp.out"
	rm -f nohup.out
	rm -f "txtreader"

test:
	go test -v ./internal/...

check:
	go test

cover:
	go test -coverprofile cp.out
	go tool cover -html=cp.out

run:
	./txtreader