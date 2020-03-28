build: proto-lint
	docker build -t testing --target test .

test: build
	docker run -v $(PWD)/artifacts:/artifacts -v /var/run/docker.sock:/var/run/docker.sock testing
	cd artifacts && curl -s https://codecov.io/bash | bash

proto-lint:
	docker run -v "$(PWD)/docs/protos:/work" uber/prototool:latest prototool lint

generate: proto-lint
	mkdir -p internal/protos
	docker run -v "$(PWD)/docs/protos:/work" -v $(PWD):/output -u `id -u $(USER)`:`id -g $(USER)` uber/prototool:latest prototool generate

ship:
	docker login --username ${DOCKER_USERNAME} --password ${DOCKER_PASSWORD}
	docker build -t syncromatics/kvetch:${VERSION} --target final .
	docker push syncromatics/kvetch:${VERSION}