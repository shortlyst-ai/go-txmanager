build: dep
	CGO_ENABLED=0 GOOS=linux go build ./...

dep:
	@echo ">> Downloading Dependencies"
	@go mod download

test: dep
	@echo ">> Running Tests"
	@env $$(cat .env.testing | xargs) go test -failfast -count=1 -p=1 -cover -covermode=atomic ./...