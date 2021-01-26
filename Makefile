# Makefile used to simplify build and testing execution

tests:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up --build tests

bench:
	@echo "Launching Benchmarks in Docker Compose"
	docker-compose -f dev-compose.yml up --build benchmarks

clean:
	@echo "Cleaning up"
	-docker-compose -f dev-compose.yml down
