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
	swagger generate server -f ./swagger/spec.yaml -s swagger/generated/restapi -m swagger/generated/models --exclude-main
	go generate ./ent

test:
	go test -race ./... -coverprofile=coverage.out

coverage:
	go tool cover -func=coverage.out

gen-repo-mock:
	@docker run -v `pwd`:/src -w /src vektra/mockery --case snake --dir swagger/repositories --output internal/mocks/repositories --outpkg repositories --all

	