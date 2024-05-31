.PHONY: all format test build

all: format test build

format:
	go fmt ./...
	
test:
	go test -v ./...

build:
	go build -o k8s-openstack-pv-exporter.exe .

helm-html:
	helm repo index . --url https://drumato.github.io/k8s-openstack-pv-exporter
	
