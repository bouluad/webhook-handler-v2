# ==============================================================================
# PROJECT AUTOMATION MAKEFILE
# Targets for the two-component Webhook Gateway (receiver and forwarder)
# ==============================================================================

# --- Configuration ---
PROJECT_NAME := webhook-gateway-project
RECEIVER_APP := receiver-app
FORWARDER_APP := forwarder-app
RECEIVER_IMAGE := webhook-receiver
FORWARDER_IMAGE := onprem-forwarder
IMAGE_TAG := latest # Change this for production releases
ACR_NAME := your-acr-name.azurecr.io # Update this to your Azure Container Registry

# --- Go Build Flags ---
# -s -w: Strips debug information and symbol tables for smaller binaries
# CGO_ENABLED=0: Creates a statically linked binary (required for Alpine/scratch Docker images)
BUILD_FLAGS := -ldflags "-s -w"
OS := linux

# ==============================================================================
# 1. SETUP AND CLEANUP
# ==============================================================================

.PHONY: tidy clean all

all: receiver forwarder docker-receiver docker-forwarder

tidy:
	@echo "--> Downloading and verifying Go dependencies..."
	go mod tidy

clean:
	@echo "--> Cleaning up binaries and Docker images..."
	rm -f $(RECEIVER_APP) $(FORWARDER_APP)
	@if docker images | grep -q $(RECEIVER_IMAGE) ; then docker rmi -f $(RECEIVER_IMAGE):$(IMAGE_TAG) || true; fi
	@if docker images | grep -q $(FORWARDER_IMAGE) ; then docker rmi -f $(FORWARDER_IMAGE):$(IMAGE_TAG) || true; fi

# ==============================================================================
# 2. BINARY BUILD (Statically Linked)
# ==============================================================================

receiver: tidy
	@echo "--> Building receiver binary ($(RECEIVER_APP)) for $(OS)..."
	GOOS=$(OS) CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(RECEIVER_APP) ./cmd/receiver

forwarder: tidy
	@echo "--> Building forwarder binary ($(FORWARDER_APP)) for $(OS)..."
	GOOS=$(OS) CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(FORWARDER_APP) ./cmd/forwarder

# ==============================================================================
# 3. DOCKER IMAGE MANAGEMENT
# ==============================================================================

docker-receiver: receiver
	@echo "--> Building Docker image for receiver ($(RECEIVER_IMAGE):$(IMAGE_TAG))..."
	docker build -t $(RECEIVER_IMAGE):$(IMAGE_TAG) -f cmd/receiver/Dockerfile .

docker-forwarder: forwarder
	@echo "--> Building Docker image for forwarder ($(FORWARDER_IMAGE):$(IMAGE_TAG))..."
	docker build -t $(FORWARDER_IMAGE):$(IMAGE_TAG) -f cmd/forwarder/Dockerfile .

push-receiver: docker-receiver
	@echo "--> Pushing receiver image to $(ACR_NAME)..."
	docker tag $(RECEIVER_IMAGE):$(IMAGE_TAG) $(ACR_NAME)/$(RECEIVER_IMAGE):$(IMAGE_TAG)
	docker push $(ACR_NAME)/$(RECEIVER_IMAGE):$(IMAGE_TAG)

push-forwarder: docker-forwarder
	@echo "--> Pushing forwarder image to $(ACR_NAME)..."
	docker tag $(FORWARDER_IMAGE):$(IMAGE_TAG) $(ACR_NAME)/$(FORWARDER_IMAGE):$(IMAGE_TAG)
	docker push $(ACR_NAME)/$(FORWARDER_IMAGE):$(IMAGE_TAG)

# ==============================================================================
# 4. TESTING AND DEVELOPMENT (Local Run)
# ==============================================================================

# NOTE: Environment variables (AZURE_SERVICE_BUS_CONN_STRING, etc.) must be 
#       exported in your shell before running these targets.

run-receiver: receiver
	@echo "--> Running receiver locally on :8080. Press Ctrl+C to stop."
	./$(RECEIVER_APP)

run-forwarder: forwarder
	@echo "--> Running forwarder locally. Press Ctrl+C to stop."
	./$(FORWARDER_APP)
