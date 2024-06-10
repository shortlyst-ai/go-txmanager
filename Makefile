build: dep
	CGO_ENABLED=0 GOOS=linux go build ./...

dep:
	@echo ">> Downloading Dependencies"
	@go mod download

test: dep
	@echo ">> Running Tests"
	@env $$(cat .env.testing | xargs) go test -failfast -count=1 -p=1 -cover -covermode=atomic ./...

test-infra-up:
	$(MAKE) test-infra-down
	@echo ">> Starting Test DB"
	docker network create go-txmanager-test
	docker run -d --rm --name go-txmanager-test-mysql --network go-txmanager-test -p 3366:3306 --env-file .env.testing mysql:5.7
	# Wait for MySQL container to be healthy
	@echo ">> Waiting for MySQL to be ready..."
	docker exec go-txmanager-test-mysql sh -c 'while ! mysqladmin ping -h localhost -u root -p"$$MYSQL_ROOT_PASSWORD" --silent; do sleep 1; done'
	@echo ">> MySQL is ready!"

test-infra-down:
	@echo ">> Shutting Down Test DB"
	@-docker kill go-txmanager-test-mysql
	@-docker network rm go-txmanager-test

