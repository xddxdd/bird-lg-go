# Basic definitions
DOCKER_USERNAME := xddxdd
ARCHITECTURES := amd64 i386 arm32v7 arm64v8 ppc64le s390x
IMAGES := frontend proxy

# General Purpose Preprocessor config
GPP_INCLUDE_DIR := include
GPP_FLAGS_U := "" "" "(" "," ")" "(" ")" "\#" ""
GPP_FLAGS_M := "\#" "\n" " " " " "\n" "(" ")"
GPP_FLAGS_EXTRA := +c "\\\n" ""
GPP_FLAGS := -I ${GPP_INCLUDE_DIR} --nostdinc -U ${GPP_FLAGS_U} -M ${GPP_FLAGS_M} ${GPP_FLAGS_EXTRA}

BUILD_ID ?= $(shell date +%Y%m%d%H%M)

define create-image-arch-target
frontend/Dockerfile.$1: frontend/template.Dockerfile
	@gpp ${GPP_FLAGS} -D ARCH_$(shell echo $1 | tr a-z A-Z) -o frontend/Dockerfile.$1 frontend/template.Dockerfile || rm -rf frontend/Dockerfile.$1

frontend/$1: frontend/Dockerfile.$1
	@if [ -f frontend/Dockerfile.$1 ]; then \
		docker build --pull --no-cache -t ${DOCKER_USERNAME}/bird-lg-go:$1-${BUILD_ID} -f frontend/Dockerfile.$1 frontend || exit 1; \
		docker push ${DOCKER_USERNAME}/bird-lg-go:$1-${BUILD_ID} || exit 1; \
		docker tag ${DOCKER_USERNAME}/bird-lg-go:$1-${BUILD_ID} ${DOCKER_USERNAME}/bird-lg-go:$1 || exit 1; \
		docker push ${DOCKER_USERNAME}/bird-lg-go:$1 || exit 1; \
	else \
		echo "Dockerfile generation failed, see error above"; \
		exit 1; \
	fi

proxy/Dockerfile.$1: proxy/template.Dockerfile
	@gpp ${GPP_FLAGS} -D ARCH_$(shell echo $1 | tr a-z A-Z) -o proxy/Dockerfile.$1 proxy/template.Dockerfile || rm -rf proxy/Dockerfile.$1

proxy/$1: proxy/Dockerfile.$1
	@if [ -f proxy/Dockerfile.$1 ]; then \
		docker build --pull --no-cache -t ${DOCKER_USERNAME}/bird-lgproxy-go:$1-${BUILD_ID} -f proxy/Dockerfile.$1 proxy || exit 1; \
		docker push ${DOCKER_USERNAME}/bird-lgproxy-go:$1-${BUILD_ID} || exit 1; \
		docker tag ${DOCKER_USERNAME}/bird-lgproxy-go:$1-${BUILD_ID} ${DOCKER_USERNAME}/bird-lgproxy-go:$1 || exit 1; \
		docker push ${DOCKER_USERNAME}/bird-lgproxy-go:$1 || exit 1; \
	else \
		echo "Dockerfile generation failed, see error above"; \
		exit 1; \
	fi

endef

$(foreach arch,${ARCHITECTURES},$(eval $(call create-image-arch-target,$(arch))))

frontend:$(foreach arch,latest ${ARCHITECTURES},frontend/${arch})

frontend/latest: frontend/amd64
	@docker tag ${DOCKER_USERNAME}/bird-lg-go:amd64-${BUILD_ID} ${DOCKER_USERNAME}/bird-lg-go:${BUILD_ID} || exit 1
	@docker push ${DOCKER_USERNAME}/bird-lg-go:${BUILD_ID} || exit 1
	@docker tag ${DOCKER_USERNAME}/bird-lg-go:amd64-${BUILD_ID} ${DOCKER_USERNAME}/bird-lg-go:latest || exit 1
	@docker push ${DOCKER_USERNAME}/bird-lg-go:latest || exit 1

proxy:$(foreach arch,latest ${ARCHITECTURES},proxy/${arch})

proxy/latest: proxy/amd64
	@docker tag ${DOCKER_USERNAME}/bird-lgproxy-go:amd64-${BUILD_ID} ${DOCKER_USERNAME}/bird-lgproxy-go:${BUILD_ID} || exit 1
	@docker push ${DOCKER_USERNAME}/bird-lgproxy-go:${BUILD_ID} || exit 1
	@docker tag ${DOCKER_USERNAME}/bird-lgproxy-go:amd64-${BUILD_ID} ${DOCKER_USERNAME}/bird-lgproxy-go:latest || exit 1
	@docker push ${DOCKER_USERNAME}/bird-lgproxy-go:latest || exit 1

.DEFAULT_GOAL := images
.DELETE_ON_ERROR:
.SECONDARY:

# Target to enable multiarch support
_crossbuild:
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes >/dev/null

dockerfiles: $(foreach image,${IMAGES},$(foreach arch,${ARCHITECTURES},$(image)/Dockerfile.$(arch)))

images: $(foreach image,${IMAGES},$(image))

clean:
	@rm -rf */Dockerfile.{$(shell echo ${ARCHITECTURES} | sed "s/ /,/g")}
