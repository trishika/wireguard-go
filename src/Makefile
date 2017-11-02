all: wireguard-go

wireguard-go: main.go $(wildcard wireguard/*.go)
	go build -o $@

clean:
	rm -f wireguard-go

cloc:
	cloc $(filter-out xchacha20.go $(wildcard *_test.go), $(wildcard *.go))

.PHONY: clean cloc
