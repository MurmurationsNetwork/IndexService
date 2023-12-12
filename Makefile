#--------------------------
# Include other Makefiles.
#--------------------------
include ./build/geoip/mk/Makefile
include ./build/index/mk/Makefile
include ./build/library/mk/Makefile
include ./build/revalidatenode/mk/Makefile
include ./build/nodecleaner/mk/Makefile
include ./build/schemaparser/mk/Makefile
include ./build/validation/mk/Makefile
include ./build/dataproxyrefresher/mk/Makefile

#--------------------------
# Set environment variables.
#--------------------------
DEPLOY_ENV ?= local

ifeq ($(DEPLOY_ENV), staging)
	ENV_FILE = test/e2e-staging-env.json
else ifeq ($(DEPLOY_ENV), pretest)
	ENV_FILE = test/e2e-pretest-env.json
else
	ENV_FILE = test/e2e-local-env.json
endif

# The TAG value is constructed based on the commit SHA.
# If running in a GitHub Actions environment, it uses the GITHUB_SHA.
# In a local environment, it uses the SHA of the HEAD commit.
TAG ?= $(shell git rev-parse --short $(if $(GITHUB_SHA),$(GITHUB_SHA),HEAD))

DOCKER_TAG_PREFIX := $(if $(filter production,$(DEPLOY_ENV)),,${DEPLOY_ENV}-)

#--------------------------
# Set up the development servers
#--------------------------
.PHONY: dev
dev:
    # There's no particular deployment order for Helm, so we use --tolerate-failures-until-deadline to
    # prevent deployment failure if the required object, such as PriorityClass, has not been deployed yet.
	export SOURCEPATH=$(PWD) && skaffold dev --tolerate-failures-until-deadline=true --port-forward

#--------------------------
# Runs the unit tests.
#--------------------------
.PHONY: test
test:
	export ENV=test && go test ./...

#--------------------------
# Run the end-to-end (E2E) tests using newman.
#--------------------------
.PHONY: newman-test
newman-test:
	newman run test/e2e-tests.json -e $(ENV_FILE) --verbose --delay-request 1000

# ---------------------------------------------------------------

docker-build-dataproxy:
	$(MAKE) -C services/dataproxy/ docker-build

docker-build-dataproxyupdater:
	$(MAKE) -C services/cronjob/dataproxyupdater/ docker-build

# ---------------------------------------------------------------

check-clean:
	@if [ -n "$(shell git status --porcelain)" ]; then \
		echo "Uncommitted changes present. Please commit them before running this command."; \
		exit 1; \
	fi

docker-tag-index: check-clean docker-build-index
	docker tag murmurations/index murmurations/$(DOCKER_TAG_PREFIX)index:$(TAG)

docker-tag-validation: check-clean docker-build-validation
	docker tag murmurations/validation murmurations/$(DOCKER_TAG_PREFIX)validation:${TAG}

docker-tag-library: check-clean docker-build-library
	docker tag murmurations/library murmurations/$(DOCKER_TAG_PREFIX)library:${TAG}

docker-tag-nodecleaner: check-clean docker-build-nodecleaner
	docker tag murmurations/nodecleaner murmurations/$(DOCKER_TAG_PREFIX)nodecleaner:${TAG}

docker-tag-schemaparser: check-clean docker-build-schemaparser
	docker tag murmurations/schemaparser murmurations/$(DOCKER_TAG_PREFIX)schemaparser:${TAG}

docker-tag-revalidatenode: check-clean docker-build-revalidatenode
	docker tag murmurations/revalidatenode murmurations/$(DOCKER_TAG_PREFIX)revalidatenode:${TAG}

docker-tag-geoip: check-clean docker-build-geoip
	docker tag murmurations/geoip murmurations/$(DOCKER_TAG_PREFIX)geoip:${TAG}

docker-tag-dataproxy: check-clean docker-build-dataproxy
	docker tag murmurations/dataproxy murmurations/$(DOCKER_TAG_PREFIX)dataproxy:${TAG}

docker-tag-dataproxyupdater: check-clean docker-build-dataproxyupdater
	docker tag murmurations/dataproxyupdater murmurations/$(DOCKER_TAG_PREFIX)dataproxyupdater:${TAG}

docker-tag-dataproxyrefresher: check-clean docker-build-dataproxyrefresher
	docker tag murmurations/dataproxyrefresher murmurations/$(DOCKER_TAG_PREFIX)dataproxyrefresher:${TAG}

# ---------------------------------------------------------------

docker-push-index: docker-tag-index
	docker push murmurations/$(DOCKER_TAG_PREFIX)index:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)index:$(TAG)

docker-push-validation: docker-tag-validation
	docker push murmurations/$(DOCKER_TAG_PREFIX)validation:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)validation:$(TAG)

docker-push-library: docker-tag-library
	docker push murmurations/$(DOCKER_TAG_PREFIX)library:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)library:$(TAG)

docker-push-nodecleaner: docker-tag-nodecleaner
	docker push murmurations/$(DOCKER_TAG_PREFIX)nodecleaner:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)nodecleaner:$(TAG)

docker-push-schemaparser: docker-tag-schemaparser
	docker push murmurations/$(DOCKER_TAG_PREFIX)schemaparser:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)schemaparser:$(TAG)

docker-push-revalidatenode: docker-tag-revalidatenode
	docker push murmurations/$(DOCKER_TAG_PREFIX)revalidatenode:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)revalidatenode:$(TAG)

docker-push-geoip: docker-tag-geoip
	docker push murmurations/$(DOCKER_TAG_PREFIX)geoip:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)geoip:$(TAG)

docker-push-dataproxy: docker-tag-dataproxy
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxy:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxy:$(TAG)

docker-push-dataproxyupdater: docker-tag-dataproxyupdater
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxyupdater:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxyupdater:$(TAG)

docker-push-dataproxyrefresher: docker-tag-dataproxyrefresher
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxyrefresher:latest
	docker push murmurations/$(DOCKER_TAG_PREFIX)dataproxyrefresher:$(TAG)

# ---------------------------------------------------------------

deploy-murmurations-core:
	helm upgrade murmurations-core ./charts/murmurations/charts/core \
	--set global.env=$(DEPLOY_ENV) --install --atomic

deploy-ingress:
	helm upgrade murmurations-ingress ./charts/murmurations/charts/ingress \
	--set global.env=$(DEPLOY_ENV) --install --atomic

deploy-mq:
	helm upgrade murmurations-mq ./charts/murmurations/charts/message-queue \
	--set global.env=$(DEPLOY_ENV) --install --atomic

deploy-index:
	helm upgrade murmurations-index ./charts/murmurations/charts/index \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)index:$(TAG) \
	--install --atomic

deploy-validation:
	helm upgrade murmurations-validation ./charts/murmurations/charts/validation \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)validation:$(TAG) \
	--install --atomic

deploy-library:
	helm upgrade murmurations-library ./charts/murmurations/charts/library \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)library:$(TAG) \
	--install --atomic

deploy-nodecleaner:
	helm upgrade murmurations-nodecleaner ./charts/murmurations/charts/nodecleaner \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)nodecleaner:$(TAG) \
	--install --atomic

deploy-schemaparser:
	helm upgrade murmurations-schemaparser ./charts/murmurations/charts/schemaparser \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)schemaparser:$(TAG) \
	--install --atomic

deploy-revalidatenode:
	helm upgrade murmurations-revalidatenode ./charts/murmurations/charts/revalidatenode \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)revalidatenode:$(TAG) \
	--install --atomic

deploy-geoip:
	helm upgrade murmurations-geoip ./charts/murmurations/charts/geoip \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)geoip:$(TAG) \
	--install --atomic

deploy-dataproxy:
	helm upgrade murmurations-dataproxy ./charts/murmurations/charts/dataproxy \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)dataproxy:$(TAG) \
	--install --atomic

deploy-dataproxyupdater:
	helm upgrade murmurations-dataproxyupdater ./charts/murmurations/charts/dataproxyupdater \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)dataproxyupdater:$(TAG) \
	--install --atomic

deploy-dataproxyrefresher:
	helm upgrade murmurations-dataproxyrefresher ./charts/murmurations/charts/dataproxyrefresher \
	--set global.env=$(DEPLOY_ENV),image=murmurations/$(DOCKER_TAG_PREFIX)dataproxyrefresher:$(TAG) \
	--install --atomic

# ---------------------------------------------------------------
# Manual Helm deployment targets for individual services with debugging.
# ---------------------------------------------------------------

# Set the specific tag and environment for manual deployments
SPECIFIC_TAG ?= latest
ENV ?= murmproto
MANUAL_DEPLOY_TARGETS = manually-deploy-murmurations-core \
                        manually-deploy-ingress \
                        manually-deploy-mq \
                        manually-deploy-index \
                        manually-deploy-validation \
                        manually-deploy-library \
                        manually-deploy-nodecleaner \
                        manually-deploy-schemaparser \
                        manually-deploy-revalidatenode \
                        manually-deploy-geoip \
                        manually-deploy-dataproxy \
                        manually-deploy-dataproxyupdater \
                        manually-deploy-dataproxyrefresher

deploy-all-services: $(MANUAL_DEPLOY_TARGETS)

manually-deploy-murmurations-core:
	helm upgrade murmurations-core ./charts/murmurations/charts/core \
		--set global.env=$(ENV) --install --atomic --debug

manually-deploy-ingress:
	helm upgrade murmurations-ingress ./charts/murmurations/charts/ingress \
	--set global.env=$(ENV) --install --atomic --debug

manually-deploy-mq:
	helm upgrade murmurations-mq ./charts/murmurations/charts/message-queue \
	--set global.env=$(ENV) --install --atomic --debug

manually-deploy-index:
	helm upgrade murmurations-index ./charts/murmurations/charts/index \
	--set global.env=$(ENV),image=murmurations/index:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-validation:
	helm upgrade murmurations-validation \
	./charts/murmurations/charts/validation \
	--set global.env=$(ENV),image=murmurations/validation:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-library:
	helm upgrade murmurations-library ./charts/murmurations/charts/library \
	--set global.env=$(ENV),image=murmurations/library:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-nodecleaner:
	helm upgrade murmurations-nodecleaner \
	./charts/murmurations/charts/nodecleaner \
	--set global.env=$(ENV),image=murmurations/nodecleaner:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-schemaparser:
	helm upgrade murmurations-schemaparser \
	./charts/murmurations/charts/schemaparser \
	--set global.env=$(ENV),image=murmurations/schemaparser:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-revalidatenode:
	helm upgrade murmurations-revalidatenode \
	./charts/murmurations/charts/revalidatenode \
	--set global.env=$(ENV),image=murmurations/revalidatenode:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-geoip:
	helm upgrade murmurations-geoip ./charts/murmurations/charts/geoip \
	--set global.env=$(ENV),image=murmurations/geoip:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-dataproxy:
	helm upgrade murmurations-dataproxy ./charts/murmurations/charts/dataproxy \
	--set global.env=$(ENV),image=murmurations/dataproxy:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-dataproxyupdater:
	helm upgrade murmurations-dataproxyupdater \
	./charts/murmurations/charts/dataproxyupdater \
	--set global.env=$(ENV),image=murmurations/dataproxyupdater:$(SPECIFIC_TAG) \
	--install --atomic --debug

manually-deploy-dataproxyrefresher:
	helm upgrade murmurations-dataproxyrefresher \
	./charts/murmurations/charts/dataproxyrefresher \
	--set global.env=$(ENV),image=murmurations/dataproxyrefresher:$(SPECIFIC_TAG) \
	--install --atomic --debug
