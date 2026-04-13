set(DEPEND_LIBS
        -Wl,--start-group
        pthread dl acc_tcp_net rt ttp_framework
        ${OCK_PROJECT_SOURCE}/output/3rdparty/googletest/lib/libgtest.a
        ${OCK_PROJECT_SOURCE}/output/3rdparty/mockcpp/lib/libmockcpp.a
        -Wl,--end-group)

file(GLOB_RECURSE UT_SOURCE_FILES *.cpp testcase/*)

add_executable(test_ttp ${UT_SOURCE_FILES})

target_compile_options(test_ttp PUBLIC -D_GNU_SOURCE -g)
target_compile_definitions(test_ttp PRIVATE BUILD_PATH="${OCK_TEST_PROJECT_BINARY}")
DFC_UT_HEADER(test_ttp)

set_target_properties(test_ttp PROPERTIES OUTPUT_NAME "test_ttp")
set_target_properties(test_ttp PROPERTIES CLEAN_DIRECT_OUTPUT 1)
link_directories(
        ${OCK_PROJECT_SOURCE}/output/lib
        ${OCK_PROJECT_SOURCE}/output/3rdparty/mockcpp/lib
        ${OCK_PROJECT_SOURCE}/output/3rdparty/googletest/lib
)

target_link_libraries(test_ttp ${DEPEND_LIBS})
target_include_directories(test_ttp PRIVATE testcase/controller/
    testcase/
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/common
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/security
    ${OCK_PROJECT_SOURCE}/src/csrc/acc_links/under_api/openssl
)

add_dependencies(test_ttp googletest mock_cpp)