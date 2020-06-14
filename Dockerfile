FROM golang:1.14 as build
ARG BUILD_FLAGS
WORKDIR /build
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go get -u golang.org/x/lint/golint
COPY . ./
RUN go vet ./...
RUN golint -set_exit_status ./...

# Build kvetch
RUN go build -o kvetch ./cmd/kvetch

# Build kvetchctl
RUN mkdir -p \
    /artifacts/linux \
    /artifacts/arm \
    /artifacts/windows \
    /artifacts/darwin
RUN GOOS=linux GOARCH=amd64 go build -ldflags "$BUILD_FLAGS" -o /artifacts/linux/kvetchctl ./cmd/kvetchctl
RUN GOOS=linux GOARCH=arm go build -ldflags "$BUILD_FLAGS" -o /artifacts/arm/kvetchctl ./cmd/kvetchctl
RUN GOOS=darwin GOARCH=amd64 go build -ldflags "$BUILD_FLAGS" -o /artifacts/darwin/kvetchctl ./cmd/kvetchctl
RUN GOOS=windows GOARCH=amd64 go build -ldflags "$BUILD_FLAGS" -o /artifacts/windows/kvetchctl.exe ./cmd/kvetchctl

# Test repo
FROM build as test
CMD go test -race -coverprofile=/artifacts/coverage.txt -covermode=atomic ./...

# Package binaries for release
FROM build AS package
RUN apt-get -y update && apt-get -y install zip
WORKDIR /artifacts
RUN cd linux && tar zcf /artifacts/linux.tar.gz kvetchctl
RUN cd arm && tar zcf /artifacts/arm.tar.gz kvetchctl
RUN cd darwin && tar zcf /artifacts/darwin.tar.gz kvetchctl
RUN cd windows && zip -r /artifacts/windows.zip kvetchctl.exe
WORKDIR /artifacts

# Final configuration for running kvetch
FROM ubuntu:18.04 as final
ENV GOMAXPROCS 128
WORKDIR /app
COPY --from=0 /build/kvetch /app/
CMD ["/app/kvetch"]
