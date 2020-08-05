.PHONY:  default build check-env docker images login push send-message

default: docker

check-env:
ifndef AWSCLI
	$(error AWSCLI is undefined)
endif

build:
	docker build \
	    --compress --force-rm --pull \
	    -t runner:latest \
	    -f ./cmd/runner/Dockerfile \
	    .
	docker build \
	    --compress --force-rm --pull \
	    -t `terraform output ambassador_repo_url`:latest \
	    -f ./cmd/ambassador/Dockerfile \
	    .
	docker build \
	    --compress --force-rm --pull \
	    -t `terraform output application_repo_url`:latest \
	    -f ./cmd/application/Dockerfile \
	    .

docker:
	docker-compose --version
	docker-compose build --pull ambassador
	docker-compose build --pull application
	docker-compose build --pull runner
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
