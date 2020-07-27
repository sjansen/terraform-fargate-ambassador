.PHONY:  default build docker images login push

default: docker

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
	aws ecr list-images \
	    --repository-name `terraform output ambassador_repo_name`
	aws ecr list-images \
	    --repository-name `terraform output application_repo_name`

login:
	aws ecr get-login-password \
	| docker login \
	    --username AWS \
	    --password-stdin \
	    `terraform output registry`

push:
	-aws ecr batch-delete-image \
	    --image-ids imageTag=latest \
	    --repository-name `terraform output ambassador_repo_name`
	docker push `terraform output ambassador_repo_url`:latest
	-aws ecr batch-delete-image \
	    --image-ids imageTag=latest \
	    --repository-name `terraform output application_repo_name`
	docker push `terraform output application_repo_url`:latest
