#!/usr/bin/env make -f
# -*- makefile -*-

deploy:
	cd terraform && terraform apply

build:
	docker rm copy-tags-from-ec2-to-ebs || echo "Error remove container"
	docker-compose -f "docker-compose.debug.yml" up -d --build && \
	docker run --name copy-tags-from-ec2-to-ebs copy-tags-from-ec2-to-ebs:latest && \
	docker cp copy-tags-from-ec2-to-ebs:/app/main.zip .

all: build deploy
