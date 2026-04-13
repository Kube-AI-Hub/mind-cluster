/**
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#define private public
#include "mem_file_system.h"
#undef private

using namespace ock::memfs;

namespace {

class TestMemFileSystem : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
protected:
    uint64_t blkSize = 10UL << 10;
    std::string sysName = "test";
};

void TestMemFileSystem::SetUp() {}

void TestMemFileSystem::TearDown()
{
    GlobalMockObject::verify();
}

void MockBmmInit()
{
    union MockerHelper {
        int32_t (MemFsBMM::*initialize)(const MemFsBMMOptions &opt) noexcept;
        int32_t (*mockInitialize)(MemFsBMM *self, const MemFsBMMOptions &opt) noexcept;
    };
    MockerHelper helper{};
    helper.initialize = &MemFsBMM::Initialize;
    auto mocker = MOCKCPP_NS::mockAPI("&MemFsBMM::Initialize", helper.mockInitialize);
    mocker.defaults().will(returnValue(0));
}

TEST_F(TestMemFileSystem, opened_file_init_falied)
{
    OpenedFile file{};
    file.allocatedFlag = true;
    
    MemFsBMM bmm{};
    auto inode = std::make_shared<MemFsInode>(0, 0, "/", InodeType::INODE_DIR, 0, bmm);
    auto ret = file.Initialize(inode, true);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFileSystem, opened_file_release_falied)
{
    OpenedFile file{};
    file.allocatedFlag = false;
    auto ret = file.Release();
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFileSystem, mem_file_system_init_no_blk)
{
    MemFileSystem memSys{ blkSize, 0UL, sysName };
    auto ret = memSys.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, mem_file_system_init_no_mem)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    if (memSys.openedFiles != nullptr) {
        delete[] memSys.openedFiles;
        memSys.openedFiles = nullptr;
    }
    auto ret = memSys.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, get_inode_with_path_empty)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    uint64_t ino = 0UL;
    uint64_t pino = 0UL;
    std::string lastToken = "";
    auto ret = memSys.GetInodeWithPath("", ino, pino, lastToken);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFileSystem, check_rename_flags_exchange_failed)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    RenameContext context{ "src", "tgt", 0 };
    auto ret = memSys.CheckRenameFlagsExchange(context);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, check_rename_flags_no_replace)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    RenameContext context{ "src", "tgt", 0 };
    auto ret = memSys.CheckRenameFlagsNoReplace(context);
    ASSERT_EQ(0, ret);

    MemFsBMM bmm;
    context.targetInode = std::make_shared<MemFsInode>(0, 0, "/", InodeType::INODE_DIR, 0, bmm);
    ret = memSys.CheckRenameFlagsNoReplace(context);
    ASSERT_EQ(-1, ret);
}
}