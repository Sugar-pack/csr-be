ifndef ${TAG}
	TAG=$$(git rev-parse --short HEAD)
endif

tag:
	echo "TAG=${TAG}" > .env

build: tag generate
	DOCKER_BUILDKIT=1 docker build --no-cache -t go_run:${TAG} --target run .

up: 
	docker-compose --env-file .env -f ./docker/docker-compose.yaml up -d

run: build up

stop:
	docker-compose --env-file .env -f ./docker/docker-compose.yaml stop

down:
	docker-compose --env-file .env -f ./docker/docker-compose.yaml down

clean:
	docker-compose --env-file .env -f ./docker/docker-compose.yaml down || true
	docker rm lc_lint || true
	docker rmi $$(docker images -q go_lint:${TAG}) || true
	docker rmi $$(docker images -q go_run:${TAG}) || true

lint:
	DOCKER_BUILDKIT=1 docker build -t go_lint:${TAG} --target lint .
	docker run -it --name lc_lint go_lint:${TAG} || true

local_run:
	go run cmd/swagger/main.go

local_lint:
	golangci-lint run --out-format tab

generate:
	rm -rf ./swagger/generated
	swagger generate server -f ./swagger/spec.yaml -s swagger/generated/restapi -m swagger/generated/models --exclude-main
	swagger generate client -f ./swagger/spec.yaml -m swagger/generated/models
	go generate ./ent

test:
	go test -race ./... -coverprofile=coverage.out -short

coverage:
	go tool cover -func=coverage.out

# to run first time required to run "make generate" before tests
# generate is in gitlab-ci file before running tests
integration-test: tag
	docker volume prune -f && \
	DOCKER_BUILDKIT=1  docker build -f Dockerfile.test --network host --no-cache -t test_go_run:${TAG} --target run . && \
	docker-compose --env-file .env -f ./docker/docker-compose.test.yaml up -d
	go test -race -v -timeout 10m ./... -run Integration
	docker-compose --env-file .env -f ./docker/docker-compose.test.yaml down

gen-repo-mock:
	@docker run -v `pwd`:/src -w /src vektra/mockery:v2.13.1 --case snake --dir swagger/repositories --output internal/mocks/repositories --outpkg repositories --all

gen-email-client-mock:
	@docker run -v `pwd`:/src -w /src vektra/mockery:v2.13.1 --case snake --dir swagger/email --output internal/mocks/email --outpkg email --all

gen-services-mocks:
	@docker run -v `pwd`:/src -w /src vektra/mockery:v2.13.1 --case snake --dir swagger/services --output internal/mocks/services --all

linux-build-mac-version:
	CGO_ENABLED=1 GOOS=linux CGO_LDFLAGS="-static" GOARCH=amd64 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ go build ./cmd/swagger/main.go

linux-build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_LDFLAGS="-static" go build ./cmd/swagger/main.go

