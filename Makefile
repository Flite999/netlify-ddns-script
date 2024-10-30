# Variables
IMAGE_NAME = flite999/netlify-ddns-script
TAG = latest

# Build the Go application
build:
    go build -o netlify-ddns-script main.go

# Build the Docker image
docker-build:
    docker build -t $(IMAGE_NAME):$(TAG) .

# Run the Docker image for testing
docker-run:
    docker run --rm $(IMAGE_NAME):$(TAG)

# Push the Docker image to Docker Hub
docker-push:
    docker push $(IMAGE_NAME):$(TAG)

# Clean up the build artifacts
clean:
    rm -f netlify-ddns-script

# Deploy to Kubernetes
deploy:
    kubectl apply -f deployment.yaml

# Build, Dockerize, and Push
all: build docker-build docker-push clean deploy