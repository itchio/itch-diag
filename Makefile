.PHONY: all run

all:
	go build -ldflags="-s -w"

run: all
	./itch-diag
