ifndef ${TAG}
	TAG=$$(git rev-parse --short HEAD)
endif

packagesToTest=$$(go list ./... | grep -v generated)

setup:
	go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.4
	go install entgo.io/ent/cmd/ent@v0.11.8
	go install github.com/vektra/mockery/v2@v2.20.2

setup_alpine:
	apk add --update --no-cache git build-base && rm -rf /var/cache/apk/*

run:
	go run ./cmd/swagger/

clean/mocks:
	find ./internal/generated/mocks/* -exec rm -rf {} \; || true

generate/mocks: clean/mocks
	mockery --all --case snake --dir ./pkg/domain --output ./internal/generated/mocks

clean/swagger:
	rm -rf ./internal/generated/swagger

generate/swagger: clean/swagger
	swagger generate server -f ./swagger.yaml -s ./internal/generated/swagger/restapi -m ./internal/generated/swagger/models --exclude-main
	swagger generate client -c ./internal/generated/swagger/client -f ./swagger.yaml -m ./internal/generated/swagger/models

clean/ent:
	find ./internal/generated/ent/* ! -name "generate.go" -exec rm -rf {} \; || true

generate/ent: clean/ent
	go run -mod=mod entgo.io/ent/cmd/ent generate --target ./internal/generated/ent ./internal/ent/schema

generate: generate/swagger generate/ent generate/mocks

clean: clean/swagger clean/ent
	rm csr coverage.out report.txt

build:
	CGO_ENABLED=0 go build -o csr ./cmd/swagger/...

lint:
	golangci-lint run --out-format tab | tee ./report.txt

test:
	go test ${packagesToTest} -race -coverprofile=coverage.out -short

coverage:
	go tool cover -func=coverage.out

coverage_total:
	go tool cover -func=coverage.out | tail -n1 | awk '{print $3}' | grep -Eo '\d+(.\d+)?'

int-test:
	DOCKER_BUILDKIT=1  docker build -f ./int-test-infra/Dockerfile.int-test --network host --no-cache -t csr:int-test --target run .
	$(MAKE) int-infra-up
	go test -v -timeout 10m ./... -run Integration
	$(MAKE) int-infra-down

int-test-without-infra:
	go test -v -p 1 -timeout 10m ./... -run Integration

build-int-image:
	docker build -t csr:int-test -f ./int-test-infra/Dockerfile.int-test .

int-infra-up:
	docker-compose -f ./int-test-infra/docker-compose.int-test.yml up -d --wait
int-infra-down:
	docker-compose -f ./int-test-infra/docker-compose.int-test.yml down

db:
	docker-compose -f ./docker-compose.yml up -d postgres


deploy_ssh:
	ssh -o "StrictHostKeyChecking=no" -i ~/.ssh/ssh_deploy -p"${deploy_ssh_port}" "${deploy_ssh_user}@${deploy_ssh_host}" 'mkdir -p /var/www/csr/${env}/'
	scp -o "StrictHostKeyChecking=no" -i ~/.ssh/ssh_deploy -P"${deploy_ssh_port}" -r ./csr "${deploy_ssh_user}@${deploy_ssh_host}:~/tmp_csr"
	scp -o "StrictHostKeyChecking=no" -i ~/.ssh/ssh_deploy -P"${deploy_ssh_port}" -r ./config.json "${deploy_ssh_user}@${deploy_ssh_host}:/var/www/csr/${env}/"
	ssh -o "StrictHostKeyChecking=no" -i ~/.ssh/ssh_deploy -p"${deploy_ssh_port}" "${deploy_ssh_user}@${deploy_ssh_host}" \
	"sudo systemctl daemon-reload && sudo service ${env}.csr stop && cp ~/tmp_csr /var/www/csr/${env}/server && sudo service ${env}.csr start"

