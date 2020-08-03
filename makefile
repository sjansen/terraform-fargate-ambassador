.PHONY:  default build docker images login push send-message

default: docker

check-env:
ifndef AWSCLI
	$(error AWSCLI is undefined)
endif

build:
	docker build \
	    --compress --force-rm --pull \
	    -t `terraform output ambassador_repo_url`:latest \
	    ./ambassador
	docker build \
	    --compress --force-rm --pull \
	    -t `terraform output application_repo_url`:latest \
	    ./application

docker:
	docker-compose --version
	docker-compose build --pull ambassador
	docker-compose build --pull application
	docker-compose up \
	    --abort-on-container-exit \
	    --exit-code-from=ambassador \
	    --force-recreate \
	    --remove-orphans

images:
	$(AWSCLI) ecr list-images \
	    --repository-name `terraform output ambassador_repo_name`
	$(AWSCLI) ecr list-images \
	    --repository-name `terraform output application_repo_name`

login: check-env
	$(AWSCLI) ecr get-login-password \
	| docker login \
	    --username AWS \
	    --password-stdin \
	    `terraform output registry`

push: check-env
	-$(AWSCLI) ecr batch-delete-image \
	    --image-ids imageTag=latest \
	    --repository-name `terraform output ambassador_repo_name`
	docker push `terraform output ambassador_repo_url`:latest
	-$(AWSCLI) ecr batch-delete-image \
	    --image-ids imageTag=latest \
	    --repository-name `terraform output application_repo_name`
	docker push `terraform output application_repo_url`:latest

send-message:
	$(AWSCLI) sqs send-message \
	    --queue-url `terraform output queue_url` \
	    --message-body 'The cake is a lie.'
