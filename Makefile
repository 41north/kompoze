.SILENT :
.PHONY : deps install build build-prod

PROJECT_NAME=$(shell basename "$(PWD)")

GO_BASE=$(shell pwd)
GO_BIN=$(GO_BASE)/bin

VERSION:=1.0.0
LDFLAGS:=-s -w -X main.VERSION=$(TAG)

all: install

deps:
	go mod download

install:
	go install ./cmd/$(PROJECT_NAME)/main.go

build:
	go build -ldflags="-X main.VERSION=$(VERSION)" -o $(GO_BIN)/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)/main.go

build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(GO_BIN)/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)/main.go

docker:
	docker build -t 41north/kompoze:"$(VERSION)" .
	docker build -t 41north/kompoze:latest .

docker-push:
	docker push 41north/kompoze:"$(VERSION)"
	docker push 41north/kompoze:latest

dist-clean:
	rm -rf dist
	rm -f docker-templates-*.tar.gz

dist: dist-clean deps
	mkdir -p dist/alpine-linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/alpine-linux/amd64/docker-templates
	mkdir -p dist/linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/linux/amd64/docker-templates
	mkdir -p dist/linux/386 && GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/linux/386/docker-templates
	mkdir -p dist/linux/armel && GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "$(LDFLAGS)" -o dist/linux/armel/docker-templates
	mkdir -p dist/linux/armhf && GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "$(LDFLAGS)" -o dist/linux/armhf/docker-templates
	mkdir -p dist/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/darwin/amd64/docker-templates

release: dist
	tar -cvzf dockerize-alpine-linux-amd64-$(TAG).tar.gz -C dist/alpine-linux/amd64 docker-templates
	tar -cvzf dockerize-linux-amd64-$(TAG).tar.gz -C dist/linux/amd64 docker-templates
	tar -cvzf dockerize-linux-386-$(TAG).tar.gz -C dist/linux/386 docker-templates
	tar -cvzf dockerize-linux-armel-$(TAG).tar.gz -C dist/linux/armel docker-templates
	tar -cvzf dockerize-linux-armhf-$(TAG).tar.gz -C dist/linux/armhf docker-templates
	tar -cvzf dockerize-darwin-amd64-$(TAG).tar.gz -C dist/darwin/amd64 docker-templates
