all:
	echo 'Provide a target: paste clean'

vendor:
	gb vendor fetch github.com/boltdb/bolt

fmt:
	find src/ -name '*.go' -exec go fmt {} ';'

build: fmt
	gb build all

paste: build
	./bin/paste

test:
	gb test -v

clean:
	rm -rf bin/ pkg/

.PHONY: paste
