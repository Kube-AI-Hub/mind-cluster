cmake_minimum_required(VERSION 3.13)
project(TTP_TEST VERSION 1.0 LANGUAGES CXX)
set(CMAKE_CXX_STANDARD 17)

add_compile_definitions(_GLIBCXX_USE_CXX11_ABI=0)

message(STATUS "build from test path")

set(OCK_PROJECT_SOURCE "${CMAKE_SOURCE_DIR}/..")
set(OCK_PROJECT_BINARY "${CMAKE_BINARY_DIR}/build_target")
set(OCK_TEST_PROJECT_BINARY "${CMAKE_BINARY_DIR}")

message(STATUS "OCK_PROJECT_SOURCE = ${OCK_PROJECT_SOURCE}")
message(STATUS "OCK_PROJECT_BINARY = ${OCK_PROJECT_BINARY}")

# 第三方库路径
set(THIRDPARTY_OUTPUT_DIR "${OCK_PROJECT_SOURCE}/output/3rdparty")

# securec 路径
set(SECUREC_OUTPUT_INCLUDE_DIR "${THIRDPARTY_OUTPUT_DIR}/libboundscheck/include")
set(SECUREC_OUTPUT_LIB_DIR "${THIRDPARTY_OUTPUT_DIR}/libboundscheck/lib")

message(STATUS "SECUREC_OUTPUT_INCLUDE_DIR = ${SECUREC_OUTPUT_INCLUDE_DIR}")

add_definitions(-DUT_ENABLED)
add_compile_options(-fprofile-arcs -ftest-coverage)
if(CMAKE_HOST_SYSTEM_PROCESSOR MATCHES "aarch64")
    add_compile_options(-march=armv8-a+crc)
endif()

add_link_options(-lgcov --coverage)

include_directories(
    ${OCK_PROJECT_SOURCE}/src
    ${OCK_PROJECT_SOURCE}/src/csrc
    ${OCK_PROJECT_SOURCE}/src/csrc/common
    ${OCK_PROJECT_SOURCE}/src/csrc/framework
    ${OCK_PROJECT_SOURCE}/src/csrc/framework/controller
    ${OCK_PROJECT_SOURCE}/src/csrc/framework/processor

    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/common
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/security
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/under_api
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/under_api/openssl

    ${SECUREC_OUTPUT_INCLUDE_DIR}

    ${OCK_PROJECT_SOURCE}/test/3rdparty/googletest/googletest/include
    ${OCK_PROJECT_SOURCE}/test/3rdparty/googletest/googlemock/include/
    ${OCK_PROJECT_SOURCE}/test/3rdparty/mockcpp/3rdparty
    ${OCK_PROJECT_SOURCE}/test/3rdparty/mockcpp/include/
    ${OCK_PROJECT_SOURCE}/output/3rdparty/googletest/include
    ${OCK_PROJECT_SOURCE}/output/3rdparty/mockcpp/include
)

link_directories(
    ${OCK_PROJECT_SOURCE}/output/lib
    ${SECUREC_OUTPUT_LIB_DIR}
    ${PROJECT_BINARY_DIR}/build_target/src/csrc/controller
    ${PROJECT_BINARY_DIR}/build_target/src/csrc/processor
)

# 链接 securec 库
link_libraries(boundscheck)

add_subdirectory(3rdparty)
add_subdirectory(llt)