.PHONY: gen-swagger
gen-swagger:
	@docker run --rm -it  -e GOPATH=$$HOME/go:/go -v $$HOME:$$HOME -w `pwd` quay.io/goswagger/swagger generate server -f ./swagger/spec.yaml -s swagger/generated/restapi -m swagger/generated/models --exclude-main