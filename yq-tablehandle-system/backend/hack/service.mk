.PHONY: run test build
run:
	cd $(BACKEND_ROOT) && go run ./app/$(SERVICE)/boot/http
test:
	cd $(BACKEND_ROOT) && go test ./app/$(SERVICE)/... ./utility/...
build:
	cd $(BACKEND_ROOT) && mkdir -p bin && go build -o bin/$(SERVICE) ./app/$(SERVICE)/boot/http
