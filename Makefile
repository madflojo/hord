# Makefile used to simplify build and testing execution

tests:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra redis
	sleep 30
	docker-compose -f dev-compose.yml up --build tests

cover:
	@echo "Generating Coverage Report"
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

bench:
	@echo "Launching Benchmarks in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra redis
	sleep 30
	docker-compose -f dev-compose.yml up --build benchmarks

clean:
	@echo "Cleaning up"
	-docker-compose -f dev-compose.yml down
