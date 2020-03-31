FROM golang:1.14 as build

WORKDIR /build

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go vet ./...

RUN go get -u golang.org/x/lint/golint

RUN golint -set_exit_status ./...

RUN go build -o kvetch ./cmd/kvetch

#testing
FROM build as test

CMD go test -race -coverprofile=/artifacts/coverage.txt -covermode=atomic ./...

# final image
FROM ubuntu:18.04 as final

ENV GOMAXPROCS 128

WORKDIR /app

COPY --from=0 /build/kvetch /app/

CMD ["/app/kvetch"]