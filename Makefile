# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Commands
DOCKER_CMD=$(shell which docker)
SWAGGER := $(DOCKER_CMD) run --rm -it -v $(HOME):$(HOME) -w $(shell pwd) quay.io/goswagger/swagger



