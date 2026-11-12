# nstchain

<p align="center">
  <img src="./kava-logo.svg" width="300" alt="nstchain Logo">
</p>

<div align="center">

[![Twitter Follow](https://img.shields.io/twitter/follow/ISTGlobal_.svg?label=Follow&style=social)](https://x.com/ISTGlobal_)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.md)

</div>

<div align="center">

</div>

## Overview
nstchain is a decentralized finance (DeFi) blockchain platform built on the Cosmos SDK, focusing on cross-chain asset management and financial innovation. It provides a comprehensive DeFi infrastructure, including lending, trading, liquidity mining, and more.

## Key Features
- **Cross-Chain DeFi**: Supports seamless transfer and trading of multi-chain assets
- **Smart Contracts**: EVM compatibility based on Ethermint
- **Liquidity Mining**: Innovative liquidity incentive mechanisms
- **Lending Platform**: Decentralized lending and collateral system
- **Price Oracle**: Reliable price data sources
- **Governance Mechanism**: Decentralized community governance
- **Auction System**: Decentralized asset auctions
- **Atomic Swaps**: Cross-chain atomic swap protocol

## Repository governance

- Branch, review, and release policy: [`docs/GIT_GOVERNANCE.md`](docs/GIT_GOVERNANCE.md)
- Release and upgrade runbook: [`docs/RELEASE.md`](docs/RELEASE.md)
- Known test and EVM issues: [`docs/KNOWN_ISSUES.md`](docs/KNOWN_ISSUES.md)
- Production source baseline: [`docs/SOURCE_BASELINE.md`](docs/SOURCE_BASELINE.md)

## Quick Start

### System Requirements

- Go 1.22+
- Git
- Make

### Installation

```bash
# 克隆代码仓库
git clone git@github.com:alicenst/nstchain-core.git
cd nstchain

# 安装（默认安装 ./cmd/nstchain 可执行文件，名称为 nstchain）
make install

# 验证安装
nstchain version
```

### Running a Local Node

```bash
# 初始化节点
nstchain init mynode --chain-id nstchain-local

# 创建测试账户
nstchain keys add mykey

# 启动节点
nstchain start
```

## Development Guide

### Running Tests

```bash
# 运行单元测试
make test

# 运行端到端测试
make test-e2e

# 运行所有测试
make test-all
```

### Building

```bash
# 构建二进制
make build

# 构建 Docker 镜像（标记为 kava/kava:local，镜像名将在后续版本改为 nstchain）
make docker-build
```

### Code Generation

```bash
# 生成 Protobuf Go 代码
make proto-gen

# 生成 Protobuf 文档
make proto-gen-doc

# 生成并合并 Swagger（输出到 client/docs/swagger-ui/swagger.yaml）
make proto-gen-swagger
```

## Core Modules

nstchain includes the following core modules:

| Module | Description |
|------|----------|
| **auction** | Decentralized auction system |
| **bep3** | Cross-chain atomic swap protocol |
| **cdp** | Collateralized debt position management |
| **committee** | Committee governance mechanism |
| **community** | Community pool management |
| **earn** | Yield aggregator |
| **evmutil** | EVM utilities and conversions |
| **hard** | Decentralized lending platform |
| **incentive** | Liquidity mining incentives |
| **issuance** | Token issuance management |
| **kavadist** | Distribution and reward system |
| **liquid** | Liquidity management |
| **pricefeed** | Price oracle |
| **router** | Cross-module routing |
| **savings** | Savings accounts |
| **swap** | Decentralized exchange |

## Contribution Guidelines
欢迎社区贡献！请提交 Issue 或 Pull Request 来讨论与改进。

### Development Environment Setup

```bash
# 代码检查
make lint

# 代码格式化
make format
```

### Submitting Code

```bash
# 创建功能分支
git checkout -b feature/your-feature-name

# 提交改动
git commit -m "feat: add new feature"

# 推送分支
git push origin feature/your-feature-name
```

## License

Copyright © nstchain Labs, Inc. All rights reserved.

This project is licensed under the [Apache v2 License](LICENSE.md)

## Community

- **Website**: [http://nstchain.us](http://nstchain.us/)
- **Twitter**: [https://x.com/ISTGlobal_](https://x.com/ISTGlobal_)

---
Note: The main branch contains features under development and may be unstable. For production environments, please use the released versions.

