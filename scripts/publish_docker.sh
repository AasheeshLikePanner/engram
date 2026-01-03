#!/bin/bash
set -e

VERSION="v0.1.1"
IMAGE_NAME="aasheeshlikepanner/engram-engine"

echo "🐳 Building Engram Engine Docker Image ${VERSION}..."

docker build -t ${IMAGE_NAME}:${VERSION} -t ${IMAGE_NAME}:latest .

echo "🚀 Pushing Image to Docker Hub..."

docker push ${IMAGE_NAME}:${VERSION}
docker push ${IMAGE_NAME}:latest

echo "✅ Published Engram Engine ${VERSION} to Docker Hub!"
