# Define variables
IMAGE_NAME=goes-api
TAG=latest

.PHONY: build local-build-deploy clean

build:
	docker build -t $(IMAGE_NAME):$(TAG) .

run:
	go run main.go

local-build-deploy:
	docker-compose up --build

clean:
	docker-compose down
	docker rmi $(IMAGE_NAME):$(TAG)
