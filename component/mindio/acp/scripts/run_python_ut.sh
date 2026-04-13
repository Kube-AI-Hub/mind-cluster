#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.

set -e
set -u

echo "========== Start Python UT =========="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ACP_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# 1. 运行build脚本
if [ -f "${ACP_DIR}/build/build.sh" ]; then
    bash "${ACP_DIR}/build/build.sh"
    if [ $? -ne 0 ]; then
        echo "error: bash build.sh fault"
        exit 1
    fi
else
    echo "error: can not find build.sh"
    exit 1
fi

# 2. 获取Python版本信息
python_version="python$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')"
echo "Python version: ${python_version}"

# 3. 安装mindio_acp包
WHL_DIR="${ACP_DIR}/python_whl/mindio_acp"
if [ ! -d "${WHL_DIR}" ]; then
    echo "error: can not find mindio_acp directory"
    exit 1
fi

cd "${WHL_DIR}"
python3 setup.py install --prefix="${ACP_DIR}/python_test"

if [ $? -ne 0 ]; then
    echo "error: mindio_acp install fault"
    exit 1
fi

# 4. 设置PYTHONPATH
SITE_PACKAGES_DIR="${ACP_DIR}/python_test/lib/${python_version}/site-packages"

if [ ! -d "${SITE_PACKAGES_DIR}" ]; then
    echo "error: can not find site-packages directory"
    exit 1
fi

# 查找mindio_acp目录
mindio_acp_dir=$(ls "${SITE_PACKAGES_DIR}" | grep mindio_acp | head -1)
if [ -z "${mindio_acp_dir}" ]; then
    echo "error: can not find mindio_acp directory in ${SITE_PACKAGES_DIR}"
    exit 1
fi

export PYTHONPATH="${SITE_PACKAGES_DIR}:${PYTHONPATH:-}"

# 5. 运行pytest单元测试
TESTS_DIR="${ACP_DIR}/python_whl/mindio_acp/tests"

cd "${TESTS_DIR}"

# 确保pytest已安装
if ! command -v pytest &> /dev/null; then
    echo "error: pytest does not installed"
    exit 1
fi

pytest --cov=mindio_acp --cov-branch --cov-report=html --cov-report=xml --junitxml=final.xml -v -s .

# 6. 拷贝htmlcov到output目录
OUTPUT_DIR="${ACP_DIR}/output"

# 检查htmlcov目录是否存在
if [ -d "${TESTS_DIR}/htmlcov" ]; then
    rm -rf "${OUTPUT_DIR}/htmlcov"
    cp -r "${TESTS_DIR}/htmlcov" "${OUTPUT_DIR}/"
else
    echo "error: can not find htmlcov directory"
fi

# 拷贝其他报告文件
if [ -f "${TESTS_DIR}/coverage.xml" ]; then
    cp "${TESTS_DIR}/coverage.xml" "${OUTPUT_DIR}/"
fi

if [ -f "${TESTS_DIR}/final.xml" ]; then
    cp "${TESTS_DIR}/final.xml" "${OUTPUT_DIR}/"
fi

echo "========== Python UT finished =========="
echo "Location of the test report: ${OUTPUT_DIR}"