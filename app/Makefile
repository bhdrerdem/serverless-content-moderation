packages := $(shell find . -name \*main.go -not -path "./node_modules/*"  | awk -F'/' '{print $$2}')
pwd := ${CURDIR}
git_root := $(shell git rev-parse --show-toplevel)

clean:
	rm -rf ./bin

install:
	npm install

build: clean
	@for package in $(packages) ; do \
  		set -e ; \
		echo Build $(pwd)/bin/$$package ; \
		mkdir -p bin/$$package ; \
		env GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -gcflags=all=-trimpath=$(git_root) -asmflags=all=-trimpath=$(git_root) -o bin/$$package/main ./$$package/ ; \
	done


deploy-dev: install build
	serverless deploy --verbose --stage dev

deploy-prod: install build
	serverless deploy --verbose --stage prod