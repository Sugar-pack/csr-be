FROM golang:1.17-alpine as build

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY internal/utils internal/utils
COPY go.mod go.sum ./

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-extldflags '-static'" -o /go cmd/swagger/main.go


FROM alpine:3.15 as run

WORKDIR /go
COPY db db
COPY --from=build /go/main ./

ENTRYPOINT [ "./main" ]


FROM golangci/golangci-lint:v1.44-alpine as lint

WORKDIR /go/src
COPY cmd cmd
COPY ent ent
COPY swagger swagger
COPY internal/utils utils
COPY .golangci.yml .golangci.yml
COPY go.mod go.sum ./

ENTRYPOINT [ "/usr/bin/golangci-lint", "run", "--out-format", "tab" ]