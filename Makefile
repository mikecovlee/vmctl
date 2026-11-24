.PHONY: build test vet check runtime dist release clean

build:
	go build -trimpath -o bin/vmctl ./cmd/vmctl

test:
	go test -race -shuffle=on -count=1 ./...

vet:
	go vet ./...

check:
	sh scripts/check.sh

runtime:
	docker build -t vmctl-runtime:local runtime

dist:
	sh scripts/dist.sh

release:
	sh scripts/release.sh "$(VERSION)"

clean:
	rm -rf bin
