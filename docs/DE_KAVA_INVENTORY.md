# De-Kava Inventory

- 盘点日期：2026-11-13
- 基线：`44d91ca01a864fd0497b113dfbaf519e389e45d5`
- 对应里程碑：`DE_KAVA_MILESTONE.md` D1
- 方法：`git ls-files`、`rg`、`go list -m all`、`go test`（仅确认已知失败包）

本清单把 Kava 相关内容分为三类：

- **必须改**：运行身份、构建路径或测试正确性依赖它，留着就会误导或失败
- **需 fork**：依赖 Kava-Labs 维护的代码，必须由 NSTChain 自己控制
- **可保留**：历史归属、许可证声明、链上 API 兼容面，改动会破坏兼容或合规

## 1. 汇总

| 项目 | 规模 | 处置 |
|---|---|---|
| module path `github.com/kava-labs/kava` | 541 个 `.go` 文件，1229 处引用 | 必须改（D3） |
| `cmd/kava` | 1 个 main 包，19 个源文件 | 必须改（D3） |
| Kava-Labs fork 依赖 | 4 个 `replace` + 1 个直接 `require` | 需 fork（D2） |
| `kava1` Bech32 前缀 | 44 个 `.go` 文件，157 处 | 测试 fixture 必须改（D4） |
| `ukava` denom | 145 个 `.go` 文件，1800 处 | 测试 fixture 必须改（D4/D5） |
| `aist` denom | 3 个 `.go` 文件（生产变量名 + 测试 fixture） | 测试 fixture 必须改（D4） |
| proto package `kava.*` | 17 个模块，`proto/kava/**` | 可保留（链上 API 兼容） |
| 品牌/文档/配置 | 141 个文件名含 `kava`，194 个非 Go 文件含 `kava` | 必须改（D6），历史 CHANGELOG 可保留 |
| CI/发布 | `CMD=kava`、`PROJECT_NAME=nstchain`、二进制名 `nstchain` | 必须改（D3/D7） |

## 2. 依赖清单

`go list -m all` 中与 Kava-Labs 相关的模块：

| 模块 | 当前版本 | 类型 | 用途 | 处置 |
|---|---|---|---|---|
| `github.com/kava-labs/kava` | 本仓库 | module path | 自身包路径 | 改为 NSTChain 自有路径（D3） |
| `github.com/kava-labs/cometbft` | `v0.37.13-kava.1` | `replace` | CometBFT 0.37 分支，含 Kava 补丁 | fork 或 vendor（D2） |
| `github.com/kava-labs/cosmos-sdk` | `v0.47.15-iavl-v1-kava.1` | `replace` | SDK 分支：backport `unsafe-reset-all`、staking transfer events、custom tally handler | fork 或 vendor（D2） |
| `github.com/kava-labs/iavl` | `v1.2.0-kava.2` | `replace` | 兼容 v0 状态的 IAVL v1 补丁 | fork 或 vendor（D2） |
| `github.com/kava-labs/ethermint` | `v0.21.0-kava-v27.0` | `replace` | EVM 分支：`min-gas-price` + `NoBaseFee`、London、EIP-712 | fork 或 vendor（D2） |
| `github.com/Kava-Labs/opendb` | `v0.0.0-20240719173129-a2f11f6d7e51` | 直接 `require` | 数据库打开与配置封装 | fork 或 vendor（D2） |

替代方案：

1. **首选**：fork 到 `nstchain` 组织，保留原 tag/commit，发布带来源说明的 NSTChain tag。`replace` 指向新路径，记录 patch 差异。
2. **备选**：vendor 进仓库。五个模块体积大、升级成本高，只在无法建仓时使用。
3. **回上游**：逐项验证 Kava 补丁是否已进入上游；已进入的依赖改回官方路径，减少自维护面。

`tests/e2e-ibc/go.mod` 是独立模块，含同样的 4 个 `replace`，D2/D3 必须同步处理。

## 3. 代码引用

### 3.1 自身 module path

- 541 个 `.go` 文件引用 `github.com/kava-labs/kava`，共 1229 处。
- 其中 258 个是 `_test.go`，283 个是生产代码。
- 覆盖 `app/`、`cmd/kava/`、`client/`、`migrate/`、`tests/`、`x/` 全部模块。
- proto 生成文件（69 个 `.pb.go`）不引用 module path，重写后无需重新生成即可编译。

处置：D3 机械替换，单独提交，不与共识/业务改动混合。

### 3.2 非自身路径的 Kava-Labs 引用

| 位置 | 引用 | 性质 | 处置 |
|---|---|---|---|
| `cmd/kava/cmd/{root,app}.go`、`cmd/kava/cmd/rocksdb/compact.go` | `github.com/Kava-Labs/opendb` | 运行依赖 | D2 |
| `go.mod` / `tests/e2e-ibc/go.mod` | 4 个 fork `replace` | 运行依赖 | D2 |
| `tests/e2e/readme.md`、`tests/e2e/runner/kvtool.go` | `kvtool` | 本地 E2E 工具 | D6 改为可选/自带说明 |
| `x/bep3/spec/*.md` | `Kava-Labs/javascript-sdk` | 上游文档链接 | 可保留，标注归属 |
| `CHANGELOG.md`、`migrate/v0_27/*.md` | Kava-Labs 历史链接 | 历史记录 | 可保留 |
| `Makefile:46` `HTTPS_GIT` | `Kava-Labs/kava.git` | 发布元数据 | D3 改为 NSTChain 仓库 |
| `client/docs/config.json` | `api.data.kava.io`、Kava 标题 | 文档站配置 | D6 |

### 3.3 模块与命令

- `cmd/kava/main.go` 与 `cmd/kava/cmd/**`：二进制入口、root command、RocksDB 工具、IAVL viewer。
- `Makefile`：`CMD ?= kava`，`PROJECT_NAME := nstchain`，`HTTPS_GIT` 指向 Kava。
- `Dockerfile` / `Dockerfile-rocksdb*`：`WORKDIR /root/kava`，产物 `/bin/kava`，`CMD ["kava"]`。
- `.github/workflows/{ci,release}.yml`：`make build-linux CMD=kava`，产物改名 `nstchain`。
- `networks/local/kavanode`、`ci/env/kava-{internal-testnet,protonet}`：本地/CI 网络配置。
- `contrib/init/io.kava.kvd.plist`：macOS 服务定义。

处置：D3 改目录与构建参数，D6 改网络/服务配置与品牌文件。

## 4. 测试债

### 4.1 `x/evmutil`

- `kava1` Bech32：`x/evmutil/types/msg_test.go`（12 处）、`x/evmutil/keeper/msg_server_test.go`、`x/evmutil/keeper/{invariants,msg_server_bep3,conversion_evm_native_bep3}_test.go`、CLI 示例。
- `aist` denom：`x/evmutil/keeper/bank_keeper_test.go` 的 fixture（大量 `sdk.NewInt64Coin("aist", ...)`）。
- 生产代码 `x/evmutil/keeper/{keeper,bank_keeper}.go` 中的 `aistKeeper` 是内部字段名，不是链上 denom；D4 只改测试与注释，不改生产逻辑。
- 当前 `go test ./x/evmutil/...` 不是全绿；CI 只跑 `./x/evmutil` 与 `./x/evmutil/migrations/v2`。

验收：fixture 改用 `nst` 前缀与 `anst` denom；`go test ./x/evmutil/...` 全绿；加入 CI。

### 4.2 `x/earn`

- `ukava`：`x/earn` 内 10 个文件使用，测试 app 配置为 `uist`，失败信息为 `invalid deposit denom`。
- `x/earn/testutil/suite.go` 同时硬编码 `bkava-kavavaloper...`、`kava:usd` 市场与 `kava` 基础资产。
- 当前 `go test ./x/earn/...` 失败，未进入 CI。

验收：测试 denom 与链配置一致；不改生产逻辑；`go test ./x/earn/...` 全绿；加入 CI。

### 4.3 其他旧命名

- `app/test_common.go`：`TestChainId = "kavatest_2221-1"`（仅测试）。
- `kava1` 还出现在 16 个生产 CLI 文件的 help/example 文本中，属品牌清理（D6），不影响测试。
- `ukava` 在 `x/earn` 之外还有 135 个文件，多数是历史测试与迁移测试；不在 D4/D5 范围，D7 守卫只约束新引入的 `kava-labs/kava` 导入，不强制全仓替换 denom。

## 5. 可保留项

| 项目 | 原因 |
|---|---|
| `proto/kava/**` 与 `package kava.*` | 链上消息/查询 API 与客户端兼容面，改名会破坏签名、gRPC 路由与历史状态 |
| `LICENSE.md`（Apache 2.0） | 合规要求，保留 |
| `CHANGELOG.md`、`migrate/` 中的 Kava 历史链接 | 历史归属与迁移记录，保留并注明来源 |
| `x/bep3/spec` 等文档中的 upstream 链接 | 第三方归属，保留 |
| `x/evmutil/keeper` 的 `aistKeeper` 内部命名 | 不影响运行与链上状态；D4 不做生产重构 |

D6 必须新增 `NOTICE` 或 `THIRD_PARTY.md`，列明 Kava 与各 fork 的来源、许可证和本地 patch。

## 6. 分类处置

### 必须改

- module path（D3）
- `cmd/kava` → `cmd/nstchain`（D3）
- Makefile/Dockerfile/CI/脚本中的 `CMD=kava`、`/bin/kava`、`HTTPS_GIT`（D3）
- `x/evmutil`、`x/earn` 测试 fixture（D4/D5）
- README、logo、CLI 示例、docs 配置、网络目录、服务定义（D6）
- CI 守卫与测试门禁（D7）

### 需 fork

- `cometbft`、`cosmos-sdk`、`iavl`、`ethermint`、`opendb`（D2）

### 可保留

- proto 包名与 gRPC 路由
- Apache 2.0 与上游归属
- 历史 CHANGELOG/迁移文档
- 测试之外的旧 denom 常量（除非影响构建或测试）

## 7. 后续任务输入

| 任务 | 本清单提供的输入 |
|---|---|
| D2 | 五个依赖的用途、版本、替代方案；两个 `go.mod` 文件 |
| D3 | 541 个文件、1229 处引用；`cmd/kava` 与构建入口 |
| D4 | `x/evmutil` 前缀/denom 位置；只改测试的边界 |
| D5 | `x/earn` 的 `ukava` 与 `kvtool` fixture |
| D6 | 品牌文件、网络目录、docs 配置、NOTICE 缺口 |
| D7 | 需纳入 CI 的包；守卫应允许 proto 与历史文档 |
| D8 | 重命名不触及 proto/chain-id；升级名 `v0.28.0` 保持不变 |
