FROM golang:1.17-alpine as build

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY go.mod go.sum ./

RUN go get -d github.com/go-swagger/go-swagger/cmd/swagger && \
    go get -d entgo.io/ent/cmd/ent && \
    go mod tidy

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-extldflags '-static'" -o /go cmd/swagger/main.go


FROM alpine:3.15 as run

WORKDIR /go
COPY --from=build /go/main ./

ENTRYPOINT [ "./main" ]


FROM golangci/golangci-lint:v1.44-alpine as lint

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY .golangci.yml .golangci.yml
COPY go.mod go.sum ./

RUN go mod tidy -compat=1.17

ENTRYPOINT [ "/usr/bin/golangci-lint", "run", "--out-format", "tab" ]