# Makefile for Mongo Adapter
include .env
export
NAME=gitlab.faza.io/go-framework/redisadapter
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMODTIDY=$(GOCMD) mod tidy
GOMODDOWNLOAD=$(GOCMD) mod download

.PHONY: only-docker
only-docker:
	-docker container rm -f gotest_redis_cnt
	@printf "Starting a container mapped to port ${GOTEST_REDIS_HOST}"
	docker run -d --rm --name gotest_redis_cnt -p ${GOTEST_REDIS_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_write_cnt
	docker run -d --rm --name gotest_redis_write_cnt -p ${GOTEST_REDIS_WRITE_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_write2_cnt
	docker run -d --rm --name gotest_redis_write2_cnt -p ${GOTEST_REDIS_WRITE2_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_read2_cnt
	docker run -d --rm --name gotest_redis_read2_cnt -p ${GOTEST_REDIS2_PORT}:6379 redis:5.0.5

.PHONY: test-full
test-full:
	-docker container rm -f gotest_redis_cnt
	$(GOMODTIDY) && $(GOMODDOWNLOAD)
	@printf "Starting a container mapped to port ${GOTEST_REDIS_HOST}"
	docker run -d --rm --name gotest_redis_cnt -p ${GOTEST_REDIS_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_write_cnt
	docker run -d --rm --name gotest_redis_write_cnt -p ${GOTEST_REDIS_WRITE_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_write2_cnt
	docker run -d --rm --name gotest_redis_write2_cnt -p ${GOTEST_REDIS_WRITE2_PORT}:6379 redis:5.0.5
	-docker container rm -f gotest_redis_read2_cnt
	docker run -d --rm --name gotest_redis_read2_cnt -p ${GOTEST_REDIS2_PORT}:6379 redis:5.0.5
	@printf "\nRunning tests..."
	@printf "\n================="
	@printf "\nTEST STARTS"
	@printf "\n=================\n"
	$(GOTEST) $(NAME)
	@printf "\n================="
	@printf "\nTEST ENDS"
	@printf "\n=================\n"
	-docker container rm -f gotest_redis_cnt
	-docker container rm -f gotest_redis_read2_cnt
	-docker container rm -f gotest_redis_write_cnt
	-docker container rm -f gotest_redis_write2_cnt

.PHONY: test-simple
test-simple:
	$(GOMODTIDY) && $(GOMODDOWNLOAD)
	$(GOTEST) $(NAME)

.PHONY: test-code
test-code:
	$(GOTEST) $(NAME)