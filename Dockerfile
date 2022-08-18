FROM golang:1.17-alpine as build

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY internal/utils internal/utils
COPY internal/config internal/config
COPY internal/db internal/db
COPY internal/logger internal/logger
COPY internal/migration internal/migration
COPY go.mod go.sum ./

RUN apk add build-base
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_LDFLAGS="-static" go build ./cmd/swagger/main.go


FROM alpine:3.15 as run

WORKDIR /go
COPY db db
COPY --from=build /go/src/main ./

ENTRYPOINT [ "./main" ]


FROM golangci/golangci-lint:v1.44-alpine as lint

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY internal/utils utils
COPY internal/utils internal/utils
COPY internal/config internal/config
COPY internal/db internal/db
COPY internal/logger internal/logger
COPY internal/migration internal/migration
COPY .golangci.yml .golangci.yml
COPY go.mod go.sum ./

ENTRYPOINT [ "/usr/bin/golangci-lint", "run", "--out-format", "tab" ]