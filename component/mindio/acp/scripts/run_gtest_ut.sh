#!/usr/bin/env bash
# ***********************************************************************
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
# script for Open Computing Kit DFS to run unit test
# version: 1.1.0
# change log:
#   2024-xx-xx: Keep original build flow but replace hdt with gtest
# ***********************************************************************
set -e

usage() {
    echo "Usage: $0 [ -h | --help ] [ -s | --skip ] [ -n | --no_collect ] [ -f | --filter ]"
    echo
    echo "Examples:"
    echo " 1. bash run_ut.sh -h"
    echo " 2. bash run_ut.sh -s"
    echo " 3. bash run_ut.sh -f TestBackupFileManager.*"
    echo " 4. bash run_ut.sh -s -f TestBackupFileManager.Initialize"
    echo
    exit 1;
}

# 获取项目根目录
PROJECT_HOME="$( cd "$( dirname "$0" )"/.. && pwd  )"
UT_SRC_PATH=${PROJECT_HOME}/test
UT_EXE_PATH=${PROJECT_HOME}/output/bin
UT_CONF_PATH=${PROJECT_HOME}/output/conf
GENERATE_DIR=${PROJECT_HOME}/test/build/gcover_report

SKIP_FLAG=""
UT_FILTER="*"
ENABLE_COLLECT="1"

# 拉取三方代码
cd ${UT_SRC_PATH}
if [[ ! -d ${UT_SRC_PATH}/3rdparty/googletest ]]; then
    echo "Trying to git clone gtest ..."
    cd ${UT_SRC_PATH}/3rdparty
    git clone https://gitcode.com/GitHub_Trending/go/googletest.git
    cd ${UT_SRC_PATH}/3rdparty/googletest
    git checkout v1.12.0
fi

if [[ ! -d ${UT_SRC_PATH}/3rdparty/mockcpp ]]; then
    echo "Trying to git clone mockcpp ..."
    cd ${UT_SRC_PATH}/3rdparty
    git clone https://gitcode.com/Ascend/mockcpp.git
    cd ${UT_SRC_PATH}/3rdparty/mockcpp
    git checkout v2.7

    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/include/mockcpp/JmpCode.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/include/mockcpp/mockcpp.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCode.cpp"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeArch.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeX64.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpCodeX86.h"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/JmpOnlyApiHook.cpp"
    dos2unix "${UT_SRC_PATH}/3rdparty/mockcpp/src/UnixCodeModifier.cpp"
    dos2unix ${UT_SRC_PATH}/3rdparty/patch/*.patch
fi

cd ${PROJECT_HOME}

# Parse the argument params
while true; do
    case "$1" in
        -s | --skip )
            SKIP_FLAG="0"
            shift ;;
        -n | --no_collect )
            ENABLE_COLLECT="0"
            shift ;;
        -f | --filter )
            UT_FILTER=$2
            shift 2
            ;;
        -h | --help )
            usage
            exit 0
            ;;
        * )
            break;;
    esac
done

if [[ "$SKIP_FLAG" == "" ]]; then
    # 第一步：构建主程序
    echo "Step 1: Building main program with build.sh..."
    if [ -f "${PROJECT_HOME}/build/build.sh" ]; then
        bash ${PROJECT_HOME}/build/build.sh -t debug --ut ON
        if [ 0 != $? ]; then
            echo "Failed to build main program!"
            exit 1
        fi
    else
        echo "Warning: ${PROJECT_HOME}/build/build.sh not found, skipping main build"
    fi
    
    # 第二步：构建测试用例
    echo "Step 2: Building test cases with CMake..."
    
    # 创建测试构建目录
    mkdir -p ${PROJECT_HOME}/test/build
    cd ${PROJECT_HOME}/test/build
    
    # 配置测试项目的CMake
    CMAKE_CMD="cmake ${PROJECT_HOME}/test \
        -DCMAKE_BUILD_TYPE=Debug \
        -DENABLE_UT=ON"
    
    echo "Running CMake command: ${CMAKE_CMD}"
    eval ${CMAKE_CMD}
    
    if [ 0 != $? ]; then
        echo "Failed to configure test project with CMake!"
        exit 1
    fi
    
    # 编译测试用例
    N_CPUS=$(nproc)
    JOBS=$((N_CPUS > 2 ? N_CPUS-2 : 1))
    make -j${JOBS}
    if [ 0 != $? ]; then
        echo "Failed to build test cases!"
        exit 1
    fi
    
    cd "${PROJECT_HOME}"
fi

# 设置权限
if [ -d "${UT_EXE_PATH}" ]; then
    chmod 550 -R ${UT_EXE_PATH} 2>/dev/null || true
fi

# 清理gcda文件
find ${PROJECT_HOME}/Build/Debug -type f -name "*.gcda" 2>/dev/null | xargs rm -f 2>/dev/null || true
find ${PROJECT_HOME}/test/build -type f -name "*.gcda" 2>/dev/null | xargs rm -f 2>/dev/null || true

# 创建结果目录
[ -d "${PROJECT_HOME}/test/build/res_xml" ] && rm -rf "${PROJECT_HOME}/test/build/res_xml"
mkdir -p "${PROJECT_HOME}/test/build/res_xml"

# 创建配置目录并复制配置文件
mkdir -p ${UT_CONF_PATH}
\cp -f ${PROJECT_HOME}/configs/memfs.conf ${UT_CONF_PATH}/
sed -i -e 's/data_block_pool_capacity_in_gb.*/data_block_pool_capacity_in_gb = 8/' ${UT_CONF_PATH}/memfs.conf

# 设置库路径
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PROJECT_HOME}/output/lib

# 运行测试
TEST_EXECUTABLE="ockio_test"

cd ${PROJECT_HOME}/test/build/
if [ -n "${TEST_EXECUTABLE}" ] && [ -f "${TEST_EXECUTABLE}" ]; then
    echo "Running tests with filter: ${UT_FILTER}"
    echo "Test executable: ${TEST_EXECUTABLE}"

    if [ ! -x "${TEST_EXECUTABLE}" ]; then
        chmod +x "${TEST_EXECUTABLE}"
    fi
    
    if ! ./${TEST_EXECUTABLE} --gtest_filter="${UT_FILTER}" --gtest_output=xml:${PROJECT_HOME}/test/build/res_xml/; then
        echo "Some tests failed. Check the report for details."
    else
        echo "All tests passed!"
    fi
else
    echo "Error: No test executable found!"
    echo "Please check if test cases were built successfully"
    echo "Contents of ${PROJECT_HOME}/test/build/:"
    ls -la ${PROJECT_HOME}/test/build/ 2>/dev/null || echo "  Directory empty or not found"
    echo "Contents of ${PROJECT_HOME}/Build/Debug/:"
    ls -la ${PROJECT_HOME}/Build/Debug/ 2>/dev/null | head -n 10
    exit 1
fi

[[ $ENABLE_COLLECT == "0" ]] && exit 0

# 检查lcov是否安装
if ! command -v lcov &> /dev/null; then
    echo "lcov not found, skipping coverage collection"
    exit 0
fi

# coverage
echo "Generating coverage report..."
rm -rf ${GENERATE_DIR}
mkdir -p ${GENERATE_DIR}

# 检查源目录是否存在
if [ ! -d "${PROJECT_HOME}/Build/Debug/src" ] && [ ! -d "${PROJECT_HOME}/test/build" ]; then
    echo "Warning: No source directories found for coverage analysis"
else
    # 收集覆盖率数据
    lcov --d ${PROJECT_HOME}/Build/Debug/src \
         --d ${PROJECT_HOME}/test/build \
         --c \
         --output-file ${GENERATE_DIR}/coverage.info \
         --rc lcov_branch_coverage=1 \
         --rc lcov_excl_br_line="LCOV_EXCL_BR_LINE|MFS_LOG*|ASSERT_*|BKG_*|HLOG_*|UFS_LOG*|LOG*|gLastErrorMessage" 2>/dev/null || true
    
    if [ -f "${GENERATE_DIR}/coverage.info" ]; then
        # 过滤不需要的路径
        GTEST_FILTER=""
        if [ -n "${GTEST_ROOT}" ]; then
            GTEST_FILTER="${GTEST_ROOT}/*"
        elif [ -n "${GTEST_INCLUDE_DIR}" ]; then
            GTEST_FILTER="$(dirname ${GTEST_INCLUDE_DIR})/*"
        fi
        
        lcov -r ${GENERATE_DIR}/coverage.info \
            "/usr/*" \
            "/opt/buildtools/*" \
            "*/test/*" \
            "*gtest*" \
            "*mockcpp*" \
            "*/src/util/*" \
            "*/src/sdk/memfs/python_sdk/*" \
            "*/src/sdk/memfs/sdk/*" \
            "*/src/sdk/memfs/server/*" \
            "*/src/sdk/memfs/common/ipc*" \
            "*/3rdparty/*" \
            "*_generated.h" \
            "*/src/memfs/common/memfs_*" \
            ${GTEST_FILTER} \
            -o ${GENERATE_DIR}/coverage.info \
            --rc lcov_branch_coverage=1 2>/dev/null || true
        
        # 生成HTML报告
        genhtml -o ${GENERATE_DIR}/result \
                ${GENERATE_DIR}/coverage.info \
                --show-details \
                --legend \
                --rc lcov_branch_coverage=1 2>/dev/null || true
        
        if [ -d "${GENERATE_DIR}/result" ]; then
            echo "Coverage report generated at: ${GENERATE_DIR}/result"
        fi
    fi
fi

# 合并测试结果XML
cd ${PROJECT_HOME}/test/build
echo '<?xml version="1.0" encoding="UTF-8"?>' > test_detail.xml

if [ -d "res_xml" ] && [ "$(ls -A res_xml 2>/dev/null)" ]; then
    echo "Merging test results..."
    
    tests_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | awk -F "tests=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
    failures_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | awk -F "failures=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
    disabled_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | awk -F "disabled=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
    errors_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | awk -F "errors=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
    time_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | awk -F "time=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}' | awk '{sum+=$1} END {print sum}')
    timestamp_val=$(cat res_xml/* 2>/dev/null | grep "<testsuites " | head -n 1 | awk -F "timestamp=" '{print $2}' | awk '{print $1}' | awk -F "\"" '{print $2}')
    
    if [ -n "${tests_val}" ]; then
        echo "<testsuites tests=\"${tests_val}\" failures=\"${failures_val}\" disabled=\"${disabled_val}\" errors=\"${errors_val}\" time=\"${time_val}\" timestamp=\"${timestamp_val}\" name=\"AllTests\">" >> test_detail.xml
        cat res_xml/* | grep -v testsuites | grep -v "xml version" >> test_detail.xml 2>/dev/null || true
        echo '</testsuites>' >> test_detail.xml
        
        if [ -d "${GENERATE_DIR}/result" ]; then
            cp -rvf test_detail.xml ${GENERATE_DIR}/result/ 2>/dev/null || true
            echo "Test results merged to: ${GENERATE_DIR}/result/test_detail.xml"
        fi
    fi
fi

echo "=========================================="
echo "Unit test execution completed!"
echo "Test results: ${PROJECT_HOME}/Build/Debug/res_xml/"
echo "Coverage report: ${GENERATE_DIR}/result/"
echo "=========================================="