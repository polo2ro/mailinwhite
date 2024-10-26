COVERAGE_EXPECTED=0
DKC=docker compose -f ./infrastructure/dev/docker-compose.yml -p app

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


install: ## Install go dependencies
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	
run: ## Run the application
	$(DKC) up -d --remove-orphans

build: ## Build the application.
	$(DKC) build

force-rebuild: ## Force rebuild all Docker images
	DOCKER_BUILDKIT=0 $(DKC) build --no-cache

down: ## Down all containers and volumes
	$(DKC) kill
	$(DKC) down -v

test_mail: ## test mail
	bash tools/test_mail.sh 
	docker logs mailinwhite-postfix

%:
	@:
