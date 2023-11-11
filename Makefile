all: cli server

GOBIN ?= $(shell go env GOPATH)/bin

build-cli:
	go build -o ./bin/kubectl-resourcecache ./main.go  

install-cli: build-cli
	cp ./bin/kubectl-resourcecache $(GOBIN)/kubectl-resourcecache

cli: install-cli

test: install-cli
	kubectl apply -f testdata/test-certissuer.yaml
	kubectl apply -f testdata/test-deploy.yaml
	kubectl resourcecache get