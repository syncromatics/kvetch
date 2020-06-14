VERSION?=$(shell gogitver)
COMMIT_HASH:=$(shell git rev-parse HEAD)
COMMIT_DATE:=$(shell git show -s --format=%cd --date=format:%Y-%m-%dT%T%z HEAD)
IMAGE?=syncromatics/kvetch
BUILD_FLAGS:=\
	-X github.com/syncromatics/kvetch/internal/cmd/kvetchctl.version=$(VERSION) \
	-X github.com/syncromatics/kvetch/internal/cmd/kvetchctl.commit=$(COMMIT_HASH) \
	-X github.com/syncromatics/kvetch/internal/cmd/kvetchctl.date=$(COMMIT_DATE)

build: proto-lint
	docker build --build-arg "BUILD_FLAGS=$(BUILD_FLAGS)" -t testing:$(VERSION) --target test .
	docker build --build-arg "BUILD_FLAGS=$(BUILD_FLAGS)" -t $(IMAGE):$(VERSION) --target final .
	docker build --build-arg "BUILD_FLAGS=$(BUILD_FLAGS)" -t package:$(VERSION) --target package .

test: build
	docker run -v $(PWD)/artifacts:/artifacts -v /var/run/docker.sock:/var/run/docker.sock testing:$(VERSION)
	cd artifacts && curl -s https://codecov.io/bash | bash

proto-lint:
	docker run -v "$(PWD)/docs/protos:/work" uber/prototool:latest prototool lint

package: build
	docker run --rm -v $$PWD:/data --entrypoint cp package:$(VERSION) -R . /data/artifacts

ship:
	docker login --username $(DOCKER_USERNAME) --password $(DOCKER_PASSWORD)
	docker push $(IMAGE):$(VERSION)

generate: proto-lint
	mkdir -p internal/protos
	docker run -v "$(PWD)/docs/protos:/work" -v $(PWD):/output -u `id -u $(USER)`:`id -g $(USER)` uber/prototool:latest prototool generate
