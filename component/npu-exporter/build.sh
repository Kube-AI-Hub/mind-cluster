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

VERSION="v7.3.0-kah"

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

TMP_DOCKERFILE="$(mktemp "${TMPDIR:-/tmp}/npu-exporter-buildx.XXXXXX.Dockerfile")"
cleanup() {
    rm -f "${TMP_DOCKERFILE}"
}
trap cleanup EXIT

cat > "${TMP_DOCKERFILE}" <<EOF
ARG GOLANG_IMAGE=${GOLANG_IMAGE}
ARG BASE_IMAGE=${BASE_IMAGE}

FROM \${GOLANG_IMAGE} AS builder
ARG TARGETARCH
ARG TARGETOS
ARG VERSION
ARG GOPROXY

WORKDIR /workspace/npu-exporter
COPY ascend-common /workspace/ascend-common
COPY npu-exporter /workspace/npu-exporter

WORKDIR /workspace/npu-exporter/cmd/npu-exporter
RUN CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \\
    CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" \\
    GOPROXY=\${GOPROXY} \\
    go build -mod=mod -buildmode=pie \\
    -ldflags "-s -extldflags=-Wl,-z,now -X huawei.com/npu-exporter/v6/versions.BuildName=npu-exporter -X huawei.com/npu-exporter/v6/versions.BuildVersion=\${VERSION}_\${TARGETOS}-\${TARGETARCH}" \\
    -o /out/npu-exporter

FROM \${BASE_IMAGE}

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \\
    usermod root -s /usr/sbin/nologin

COPY --from=builder /out/npu-exporter /usr/local/bin/npu-exporter
COPY npu-exporter/build/metricConfiguration.json /usr/local/metricConfiguration.json
COPY npu-exporter/build/pluginConfiguration.json /usr/local/pluginConfiguration.json

RUN chown root:root /usr/local/bin/npu-exporter && \\
    chmod 750 -R /home/HwHiAiUser && \\
    chmod 550 /usr/local/bin/ && \\
    chmod 500 /usr/local/bin/npu-exporter && \\
    chmod 440 /usr/local/metricConfiguration.json && \\
    chmod 440 /usr/local/pluginConfiguration.json && \\
    echo 'umask 027' >> /etc/profile && \\
    echo 'source /etc/profile' >> ~/.bashrc

ENV LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi

CMD ["/usr/local/bin/npu-exporter"]
EOF

docker buildx build \
    --file "${TMP_DOCKERFILE}" \
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
