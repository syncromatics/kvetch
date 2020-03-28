lint:
	docker run -v "$(PWD)/docs/protos:/work" uber/prototool:latest prototool lint
