# De-Kava Dependency Takeover

- 日期：2026-11-13
- 对应里程碑：`DE_KAVA_MILESTONE.md` D2
- 状态：方案与源信息已固定；仓库创建等待 `nstchain` 组织写权限

## 1. 目标

把当前指向 `kava-labs` / `Kava-Labs` 的五个依赖迁到 NSTChain 控制的仓库，固定来源、记录 patch 差异，保证上游仓库不可访问时仍可构建。

## 2. 当前依赖

| 模块 | 当前目标 | Tag / Commit | 用途 |
|---|---|---|---|
| `github.com/cometbft/cometbft` | `github.com/kava-labs/cometbft` | `v0.37.13-kava.1` / `cba1dbec87eae30d5338717680a95f04a38410ae` | 共识与节点 |
| `github.com/cosmos/cosmos-sdk` | `github.com/kava-labs/cosmos-sdk` | `v0.47.15-iavl-v1-kava.1` / `adafbee3c7596477a1aa2d6560a961445d81dac6` | SDK：`unsafe-reset-all`、staking transfer events、custom tally handler |
| `github.com/cosmos/iavl` | `github.com/kava-labs/iavl` | `v1.2.0-kava.2` / `85bc7644a7fd55fc01bf0a3de61df9f4d5cbf402` | 兼容 v0 状态的 IAVL v1 |
| `github.com/evmos/ethermint` | `github.com/kava-labs/ethermint` | `v0.21.0-kava-v27.0` / `113fd5fdb8ed49c2d3b041a96f45ba6670d1e56b` | `min-gas-price` + `NoBaseFee`、London、EIP-712 |
| `github.com/Kava-Labs/opendb` | 直接 `require` | `v0.0.0-20240719173129-a2f11f6d7e51` / `a2f11f6d` | 数据库打开与配置封装 |

`tests/e2e-ibc/go.mod` 有同样的四个 `replace`，必须同步修改。

## 3. 迁移选项

### 首选：fork 到 nstchain 组织

对每个上游仓库：

```bash
gh repo fork <upstream> --org nstchain --clone=false
```

然后按原 tag 创建 NSTChain tag（保留来源信息），例如：

```bash
git clone git@github.com:nstchain/cometbft.git
cd cometbft
git checkout cba1dbec87eae30d5338717680a95f04a38410ae
git tag -a nst-v0.37.13-kava.1 -m "fork of kava-labs/cometbft v0.37.13-kava.1"
git push origin nst-v0.37.13-kava.1
```

`go.mod` 改为：

```go
replace (
	github.com/cometbft/cometbft => github.com/nstchain/cometbft nst-v0.37.13-kava.1
	github.com/cosmos/cosmos-sdk => github.com/nstchain/cosmos-sdk nst-v0.47.15-iavl-v1-kava.1
	github.com/cosmos/iavl => github.com/nstchain/iavl nst-v1.2.0-kava.2
	github.com/evmos/ethermint => github.com/nstchain/ethermint nst-v0.21.0-kava-v27.0
)
```

`opendb` 改为 `github.com/nstchain/opendb v0.0.0-...` 或带 tag 的版本。

### 备选：vendor

`go mod vendor` 会把五个模块连同其余依赖复制进仓库，体积大、升级路径差。只在无法创建组织仓库时使用。

### 回上游

逐项验证 Kava patch 是否已进入官方仓库。已进入的依赖直接改回官方路径，减少自维护面。

## 4. 执行步骤

1. 在 `nstchain` 组织创建五个 fork，固定上表 commit。
2. 记录每个 fork 相对上游的 patch 差异（`git diff <upstream-tag>..<fork>`），写入 `docs/forks/<module>.md`。
3. 更新根 `go.mod` 与 `tests/e2e-ibc/go.mod` 的 `replace`/`require`。
4. `go mod tidy`，确认 `go.sum` 只增加新路径。
5. 更新 `NOTICE` 的 fork 清单与新仓库地址。
6. 更新 `scripts/check-dekava.sh` 的 allowlist。
7. 运行验证清单。

## 5. 验证清单

```bash
go mod verify
go build ./cmd/nstchain
make ci
go list -m all | grep -i kava
go list -m all | grep -i nstchain
```

离线构建验证：

```bash
GOPROXY=off GOFLAGS=-mod=mod go build ./cmd/nstchain
```

要求：

- 构建与 `make ci` 通过
- `go list -m all` 不再出现 `github.com/kava-labs/{cometbft,cosmos-sdk,iavl,ethermint}` 与 `github.com/Kava-Labs/opendb`
- 上游仓库不可访问时仍能构建
- `NOTICE`、`go.mod`、守卫白名单三处一致

## 6. 阻塞

创建仓库需要 `nstchain` 组织的写权限或 Alice 的 GitHub API 凭据。当前本机 `gh` 登录为 `ShiningRay`（对该仓库只有 pull），SSH 为 `alicenst`（可推送已有仓库，不能建仓）。拿到权限后按第 4 节执行。
