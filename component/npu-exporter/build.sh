#!/bin/bash
set -eu -o pipefail

# Multi-arch image build via docker buildx. Requires: docker buildx, and for --push:
#   docker login watering-ai-registry.cn-shanghai.cr.aliyuncs.com
# Override: PLATFORMS=... REGISTRY=... IMG_NAME=... VERSION=... DOCKER_PUSH=0
# Single platform local load: PLATFORMS=linux/amd64 DOCKER_PUSH=0 ./build.sh

REGISTRY="${REGISTRY:-watering-ai-registry.cn-shanghai.cr.aliyuncs.com/kube-ai-hub}"
IMG_NAME="${IMG_NAME:-npu-exporter}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
BUILDER_NAME="${BUILDER_NAME:-npu-exporter-builder}"
GOLANG_IMAGE="${GOLANG_IMAGE:-watering-ai-registry.cn-shanghai.cr.aliyuncs.com/kube-ai-hub/golang:1.25.5-bookworm}"
BASE_IMAGE="${BASE_IMAGE:-watering-ai-registry.cn-shanghai.cr.aliyuncs.com/kube-ai-hub/ubuntu:22.04}"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTEXT_DIR="$(cd "${ROOT_DIR}/.." && pwd)"

VERSION="v7.3.1-kah"

IMG_TAG="${REGISTRY}/${IMG_NAME}:${VERSION}"

DOCKER_BUILDX_OUTPUT="--push"
if [[ "${PLATFORMS}" != *","* ]]; then
    if [[ "${DOCKER_PUSH:-1}" != "1" ]]; then
        DOCKER_BUILDX_OUTPUT="--load"
    fi
fi

echo "Building npu-exporter image: ${IMG_TAG}"
echo "Platforms: ${PLATFORMS}"
echo "Version: ${VERSION}"

if ! docker buildx version &>/dev/null; then
    echo "Error: docker buildx is required"
    exit 1
fi

if ! docker buildx inspect "${BUILDER_NAME}" &>/dev/null; then
    echo "Creating buildx builder: ${BUILDER_NAME}"
    docker buildx create --name "${BUILDER_NAME}" --use
else
    docker buildx use "${BUILDER_NAME}"
fi

docker buildx inspect --bootstrap >/dev/null

DOCKERFILE="${ROOT_DIR}/build/Dockerfile.buildx"
if [[ ! -f "${DOCKERFILE}" ]]; then
    echo "Error: missing ${DOCKERFILE}"
    exit 1
fi

docker buildx build \
    --file "${DOCKERFILE}" \
    --platform "${PLATFORMS}" \
    --tag "${IMG_TAG}" \
    --build-arg GOLANG_IMAGE="${GOLANG_IMAGE}" \
    --build-arg BASE_IMAGE="${BASE_IMAGE}" \
    --build-arg GOPROXY="${GOPROXY}" \
    --build-arg VERSION="${VERSION}" \
    ${DOCKER_BUILDX_OUTPUT} \
    "${CONTEXT_DIR}"

if [[ "${DOCKER_BUILDX_OUTPUT}" == "--push" ]]; then
    echo "Successfully built and pushed: ${IMG_TAG}"
else
    echo "Successfully built (loaded locally): ${IMG_TAG}"
fi
