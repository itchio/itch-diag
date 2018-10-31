.PHONY: all run

all:
	go build -ldflags="-s -w"

run: all
	./itch-diag

# You can try to run this on your machine, but if you don't have our
# code signing certificate and private key, the chances are slim...
sign: all
	signtool.exe sign //v //s MY //n "itch corp." //fd sha256 //tr "http://timestamp.comodoca.com" //td sha256 ./itch-diag.exe
