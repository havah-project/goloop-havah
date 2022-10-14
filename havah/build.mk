#
#  Makefile for ICON2
#

GOCHAIN_HAVAH_IMAGE = goloop/gochain-havah:$(GL_TAG)
GOCHAIN_HAVAH_DOCKER_DIR = $(BUILD_DIR)/gochain-havah

GOLOOP_HAVAH_IMAGE = goloop-havah:$(GL_TAG)
GOLOOP_HAVAH_DOCKER_DIR = $(BUILD_DIR)/goloop-havah

gochain-havah-image: base-image-java gorun-gochain javarun-javaexec
	@ echo "[#] Building image $(GOCHAIN_HAVAH_IMAGE) for $(GL_VERSION)"
	@ \
	rm -rf $(GOCHAIN_HAVAH_IMAGE); \
	BIN_DIR=$(CROSSBIN_ROOT)-$$(docker inspect $(BASE_JAVA_IMAGE) --format "{{.Os}}-{{.Architecture}}") \
	IMAGE_BASE=$(BASE_JAVA_IMAGE) \
	GOCHAIN_HAVAH_VERSION=$(GL_VERSION) \
	GOBUILD_TAGS="$(GOBUILD_TAGS)" \
	$(BUILD_ROOT)/docker/gochain-havah/update.sh $(GOCHAIN_HAVAH_IMAGE) $(BUILD_ROOT) $(GOCHAIN_HAVAH_DOCKER_DIR)

goloop-havah-image: base-image-java gorun-goloop javarun-javaexec
	@ echo "[#] Building image $(GOLOOP_HAVAH_IMAGE) for $(GL_VERSION)"
	@ \
	rm -rf $(GOLOOP_HAVAH_DOCKER_DIR); \
	BIN_DIR=$(CROSSBIN_ROOT)-$$(docker inspect $(BASE_JAVA_IMAGE) --format "{{.Os}}-{{.Architecture}}") \
	IMAGE_BASE=$(BASE_JAVA_IMAGE) \
	GOLOOP_HAVAH_VERSION=$(GL_VERSION) \
	GOBUILD_TAGS="$(GOBUILD_TAGS)" \
	$(BUILD_ROOT)/docker/goloop-havah/update.sh $(GOLOOP_HAVAH_IMAGE) $(BUILD_ROOT) $(GOLOOP_HAVAH_DOCKER_DIR)
