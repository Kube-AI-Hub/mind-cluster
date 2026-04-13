cmake_minimum_required(VERSION 3.14.1)
project(OCKIO_TEST VERSION 1.0 LANGUAGES CXX)
set(CMAKE_CXX_STANDARD 11)

add_subdirectory(3rdparty)

find_package(Python3 COMPONENTS Interpreter Development REQUIRED)

set(DFS_PROJECT_PATH "${CMAKE_CURRENT_LIST_DIR}/..")
set(DFS_PROJECT_BINARY "${DFS_PROJECT_PATH}/Build/Debug")
set(DFS_PROJECT_SOURCE "${DFS_PROJECT_PATH}/src")
set(DFS_PROJECT_3RD_SOURCE "${DFS_PROJECT_PATH}/3rdparty")
set(DFS_PROJECT_OUTPUT "${DFS_PROJECT_PATH}/output")
set(DFS_PROJECT_3RD_BINARY "${DFS_PROJECT_OUTPUT}/3rdparty")

INCLUDE_DIRECTORIES(
        ${DFS_PROJECT_PATH}/src/include
        ${DFS_PROJECT_PATH}/src/background
        ${DFS_PROJECT_PATH}/src/background/backup
        ${DFS_PROJECT_PATH}/src/common
        ${DFS_PROJECT_PATH}/src/memfs
        ${DFS_PROJECT_PATH}/src/memfs/bmm
        ${DFS_PROJECT_PATH}/src/memfs/common
        ${DFS_PROJECT_PATH}/src/memfs/fs
        ${DFS_PROJECT_PATH}/src/sdk/memfs/python_sdk
        ${DFS_PROJECT_PATH}/src/sdk/memfs/common
        ${DFS_PROJECT_PATH}/src/sdk/memfs/sdk
        ${DFS_PROJECT_PATH}/src/sdk/memfs/server
        ${DFS_PROJECT_PATH}/src/ufs
        ${DFS_PROJECT_PATH}/test/sdk/mock
        ${DFS_PROJECT_PATH}/output/include
        ${DFS_PROJECT_BINARY}/src
        ${Python3_INCLUDE_DIRS}
        ${DFS_PROJECT_3RD_BINARY}/ubs-comm/include/hcom
        ${DFS_PROJECT_3RD_BINARY}/libboundscheck/include
        ${DFS_PROJECT_3RD_BINARY}/spdlog/include
        ${DFS_PROJECT_PATH}/test/3rdparty/googletest/googletest/include
        ${DFS_PROJECT_PATH}/test/3rdparty/googletest/googlemock/include/
        ${DFS_PROJECT_PATH}/test/3rdparty/mockcpp/3rdparty
        ${DFS_PROJECT_PATH}/test/3rdparty/mockcpp/include/
        ${DFS_PROJECT_PATH}/output/3rdparty/googletest/include
        ${DFS_PROJECT_PATH}/output/3rdparty/mockcpp/include
)

LINK_DIRECTORIES(
        ${DFS_PROJECT_OUTPUT}/lib
        ${DFS_PROJECT_3RD_BINARY}/ubs-comm/lib
        ${DFS_PROJECT_3RD_BINARY}/libboundscheck/lib
        ${DFS_PROJECT_3RD_BINARY}/spdlog/lib64
)

add_compile_options(-g -D_FILE_OFFSET_BITS=64 -fprofile-arcs -ftest-coverage -DUT_ENABLED)
add_link_options(-lgcov --coverage)

find_program(CCACHE_FOUND ccache)
if(CCACHE_FOUND)
    set(CMAKE_CXX_COMPILER_LAUNCHER ccache)
endif(CCACHE_FOUND)

AUX_SOURCE_DIRECTORY(unit_test TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/background TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/background/backup TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/common TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/memfs TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/sdk/memfs/sdk TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/sdk/memfs/common TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/sdk/memfs/server TEST_SRCS)
AUX_SOURCE_DIRECTORY(unit_test/ufs TEST_SRCS)

# 生成能被find_test_executable()找到的测试可执行文件
ADD_EXECUTABLE(ockio_test ${TEST_SRCS})
DFC_UT_HEADER(ockio_test)

# 链接GTest和其他库
TARGET_LINK_LIBRARIES(ockio_test
        -Wl,--start-group
        pthread dl rt bdm_static_tool libboundscheck.a hcom_static
        ock_memfs_core fs_bmm fs_core fs_common
        fs_common fs_server fs_sdk bdm libacp_c2python_api.a
        ${DFS_PROJECT_PATH}/output/3rdparty/googletest/lib/libgtest.a
        ${DFS_PROJECT_PATH}/output/3rdparty/mockcpp/lib/libmockcpp.a
        -Wl,--end-group)

add_dependencies(ockio_test googletest mock_cpp)

# 设置输出路径
set_target_properties(ockio_test PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
)