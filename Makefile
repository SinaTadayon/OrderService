.DEFAULT_GOAL := greeting
.PHONY: test test-docker compose-down compose compose-debug image-dev image-prd

greeting:
	@echo "commands: test test-docker compose-down compose compose-debug image-dev image-prd"

test-docker: compose test compose-down

test:
ifndef PORT
	$(error PORT is not set)
endif
	cd ./src && go clean -testcache && export PORT=$(PORT) && export APP_ENV=dev && go test -v ./... && cd ..

compose-down:
ifndef PORT
	$(error PORT is not set)
endif
	PORT=$(PORT) docker-compose kill
	PORT=$(PORT) docker-compose rm -f

compose:
ifeq (,$(wildcard src/.env))
	$(error env file not found)
endif
ifndef PORT
	$(error PORT is not set)
endif
ifndef TAG
	$(error TAG is not set)
endif
	@echo building docker image $(TAG)
	$(eval DOCKERIPADDR="$(shell ip -4 addr show scope global dev docker0 | grep inet | awk '{print $$2}' | cut -d / -f 1)")
	PORT=$(PORT) DOCKERIP=$(DOCKERIPADDR) docker-compose kill
	sleep 2
	PORT=$(PORT) docker build -t $(TAG) -f Dockerfile .
	PORT=$(PORT) DOCKERIP=$(DOCKERIPADDR) docker-compose up -d

compose-debug:
ifeq (,$(wildcard src/.env))
	$(error env file not found)
endif
ifndef PORT
	$(error PORT is not set)
endif
ifndef TAG
	$(error TAG is not set)
endif
	@echo building docker image $(TAG)
	$(eval DOCKERIPADDR="$(shell ip -4 addr show scope global dev docker0 | grep inet | awk '{print $$2}' | cut -d / -f 1)")
	PORT=$(PORT) DOCKERIP=$(DOCKERIPADDR) docker-compose kill
	sleep 2
	PORT=$(PORT) docker build -t $(TAG) -f Dockerfile_dev .
	PORT=$(PORT) DOCKERIP=$(DOCKERIPADDR) docker-compose up -d

image-dev:
ifndef PORT
	$(error PORT is not set)
endif
ifndef TAG
	$(error TAG is not set)
endif
	PORT=$(PORT) docker build -t `echo $(TAG) | sed -s /\//-/g` -f Dockerfile_dev .

image-stg:
ifndef PORT
	$(error PORT is not set)
endif
ifndef IMAGE_NAME
	$(error IMAGE_NAME is not set)
endif
	PORT=$(PORT) docker build -t $(IMAGE_NAME):staging -f Dockerfile_stg .
	docker tag $(IMAGE_NAME):staging registry.faza.io/$(IMAGE_NAME)/$(IMAGE_NAME):staging

image-prd:
ifndef PORT
	$(error PORT is not set)
endif
ifndef IMAGE_NAME
	$(error IMAGE_NAME is not set)
endif
	PORT=$(PORT) docker build -t $(IMAGE_NAME):master -f Dockerfile .
	docker tag $(IMAGE_NAME):master registry.faza.io/$(IMAGE_NAME)/$(IMAGE_NAME):master
