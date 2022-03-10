ifndef ${TAG}
	TAG=$$(git rev-parse --short HEAD)
endif

tag:
	echo "TAG=${TAG}" > .env

build: tag
	DOCKER_BUILDKIT=1 docker build -t go_run:${TAG} --target run .

up: 
	docker-compose --env-file .env -f ./docker/docker-compose.yaml up -d

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



	