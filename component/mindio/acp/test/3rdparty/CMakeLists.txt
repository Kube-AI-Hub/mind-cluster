include(ExternalProject)
get_filename_component(PROJECT_3RDPARTY_BIN_DIR 
    ${PROJECT_SOURCE_DIR}/../output/3rdparty 
    ABSOLUTE)
set(THIRD_PARTY_DIR ${PROJECT_SOURCE_DIR}/3rdparty)
set(GTEST_INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/googletest)
set(MOCKCPP_INSTALL_DIR ${PROJECT_3RDPARTY_BIN_DIR}/mockcpp)
message(STATUS "PROJECT_3RDPARTY_BIN_DIR:${PROJECT_3RDPARTY_BIN_DIR}")
message(STATUS "THIRD_PARTY_DIR:${THIRD_PARTY_DIR}")
message(STATUS "GTEST_INSTALL_DIR:${GTEST_INSTALL_DIR}")
message(STATUS "MOCKCPP_INSTALL_DIR:${MOCKCPP_INSTALL_DIR}")

set(THIRD_PARTY_CMAKE_ARGS
        -DCMAKE_BUILD_TYPE=DEBUG
        -DBUILD_SHARED_LIBS=OFF)

######################################################################
# Google Test and Google Mock
######################################################################
set(GTEST_CMAKE_ARGS
        ${THIRD_PARTY_CMAKE_ARGS}
        -DCMAKE_INSTALL_LIBDIR=${GTEST_INSTALL_DIR}/lib
        -DCMAKE_INSTALL_BINDIR=${GTEST_INSTALL_DIR}/bin
        -DCMAKE_INSTALL_PREFIX=${GTEST_INSTALL_DIR}/
        -DINSTALL_GTEST=ON
        -DBUILD_SHARED_LIBS=OFF
        -Dgmock_build_test=OFF
        -Dgtest_build_tests=OFF
        -Dgtest_build_samples=OFF)

externalproject_add(googletest
        SOURCE_DIR ${THIRD_PARTY_DIR}/googletest/
        CMAKE_ARGS ${GTEST_CMAKE_ARGS}
        BUILD_ALWAYS OFF
        LOG_DOWNLOAD On
        LOG_CONFIGURE On
        LOG_BUILD On
        LOG_INSTALL On)

macro(TARGET_GTEST_HEADER target)
    add_dependencies(${target} googletest)
    target_include_directories(${target} PUBLIC ${PROJECT_3RDPARTY_BIN_DIR}/googletest/include)
    target_link_directories(${target} PUBLIC ${PROJECT_3RDPARTY_BIN_DIR}/googletest/lib)
    target_link_libraries(${target} ${PROJECT_3RDPARTY_BIN_DIR}/googletest/lib/libgtest.a)
endmacro()

######################################################################
# Mockcpp
######################################################################
set(MOCKCPP_CMAKE_ARGS
        ${THIRD_PARTY_CMAKE_ARGS}
        -DCMAKE_INSTALL_LIBDIR=${MOCKCPP_INSTALL_DIR}/lib
        -DCMAKE_INSTALL_BINDIR=${MOCKCPP_INSTALL_DIR}/bin
        -DCMAKE_INSTALL_PREFIX=${MOCKCPP_INSTALL_DIR}/)

set(MOCKCPP_PATCH ${PROJECT_SOURCE_DIR}/3rdparty/patch/mockcpp_support_arm64.patch)

externalproject_add(mock_cpp
        SOURCE_DIR ${THIRD_PARTY_DIR}/mockcpp/
        CMAKE_ARGS ${MOCKCPP_CMAKE_ARGS}
        PATCH_COMMAND patch -p1 --forward -i ${MOCKCPP_PATCH} WORKING_DIRECTORY <SOURCE_DIR>
        BUILD_ALWAYS OFF
        LOG_DOWNLOAD On
        LOG_CONFIGURE On
        LOG_BUILD On
        LOG_INSTALL On)

macro(TARGET_MOCKCPP_HEADER target)
    add_dependencies(${target} mock_cpp)
    target_include_directories(${target} PUBLIC ${PROJECT_3RDPARTY_BIN_DIR}/mockcpp/include)
    target_link_directories(${target} PUBLIC ${PROJECT_3RDPARTY_BIN_DIR}/mockcpp/lib)
    target_link_libraries(${target} ${PROJECT_3RDPARTY_BIN_DIR}/mockcpp/lib/libmockcpp.a)
endmacro()

macro(DFC_UT_HEADER target)
    add_dependencies(${target} googletest)
    add_dependencies(${target} mock_cpp)
endmacro()

