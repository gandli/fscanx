# fscanx (gandli/fscanx) 审计白皮书

- **审计日期**: 2026-07-20
- **审计对象**: `gandli/fscanx`(fork of `killmonday/fscanx`)
- **Commit**: `38dfa84` (HEAD, master)
- **审计模式**: full(聚焦本 fork 责任区 + 发布/CI 安全面)
- **审计工具**: `go build` / `go vet` / 静态密钥与路径扫描 / CI 配置审查 / 依赖边界检查

## 综合评分

**76 / 100 · C+**

> 主程序可编译、发布流水线可用、Win7 兼容目标达成。失分主要来自:CI action 未 SHA pin(supply-chain)、win7 workflow 内置 UPX 集成已知不兼容(功能缺陷)、上游 examples 编译损坏(代码质量信号)、testdata 私钥入库、gitignore 不严、版本一致性差。

## 技术债估算

| 维度 | 估算 |
|---|---|
| 修复 P0 | 0 项(无阻断) |
| 修复 P1 | 约 4-6 小时(5 项) |
| 修复 P2 | 约 2-3 小时(6 项) |
| 上游代码治理 | Not assessed(第三方 inherited,73k LOC,超出本次范围) |

## 发现清单(P0 > P1 > P2)

### P0(阻断) — 0 项
无可阻断问题。主程序可编译、发布流水线可用、无真实凭证泄露。

### P1(严重) — 5 项

#### P1-1 · CI Action 未 SHA pin(supply-chain)
- **文件**: `.github/workflows/release.yml:17,22,27,36`; `build-win7.yml:30,35,43,48,55`; `build.yml:23,28,33,41`
- **问题**: 所有 `uses:` 用浮动 tag(`@v3`/`@v4`/`@v5`/`@v6`/`@v3`),非 40-char SHA。GitHub tag 可被维护者移动到恶意 commit,构成供应链攻击面。
- **失败场景**: `actions/checkout` 或 `goreleaser-action` 的 tag 被劫持/误移,CI 执行恶意代码,泄露 `GITHUB_TOKEN` 或注入后门二进制。
- **最小修复**: 将每个 action 替换为对应 tag 的 commit SHA,保留 tag 注释。
- **回归测试建议**: 加 `actionlint` 或手动 grep 校验所有 `uses:` 含 40-char SHA。
- **工作量**: 30 min

#### P1-2 · win7 workflow 内置 UPX 集成已知不兼容(功能缺陷)
- **文件**: `.github/conf/.goreleaser.win7.yml:24-31`(upx 块 `enabled: true`); `.github/workflows/build-win7.yml:48`(goreleaser-action `@v6` + `~> v1`)
- **问题**: 该配置启用 goreleaser v1 内置 UPX 集成,但本仓库已实证 goreleaser v1.26.2 内置 UPX 与现代 UPX(4.2.4/5.2.0)不兼容(报 `invalid option -a`)。`build-win7.yml` 用浮动 `~> v1` 会漂移到不兼容版本,Win7 构建必失败。
- **失败场景**: 用户手动 dispatch `build-win7` 时,UPX 步骤报错,产物缺失。
- **最小修复**: win7 配置也关掉内置 UPX(`enabled: false`),手动 `upx` 压缩;或将 win7 workflow 也改为 `goreleaser build` + 手动 `upx`(与 release.yml 一致)。同时把 `version: '~> v1'` 锁 `v1.26.2`。
- **回归测试建议**: 在 fork 跑一次 `build-win7` dispatch 验证产物。
- **工作量**: 20 min

#### P1-3 · 上游 examples 编译损坏(代码质量信号)
- **文件**: `mylib/finger/cmd/nmap/nmap.go:81,142`; `mylib/finger/cmd/test/main.go:48,95,138,284`; `mylib/finger/cmd/engine/example.go:161`; `mylib/finger/wappalyzer/examples/main.go:19`; `mylib/grdp/plugin/rdpgfx/rdpgfx.go:21,24`
- **问题**: `go build ./...` 在多个 examples/cmd 报 undefined(`fingers`、`wappalyzer.New`、`c.w`)。这些是上游 inherited 代码,主程序(`.`)不受影响,但说明上游代码存在未维护的 broken references。
- **失败场景**: 若未来有人把某个 example 纳入构建路径,CI 会失败;也反映上游维护质量。
- **最小修复**: 本 fork 范围——在 CI 加 `go build .`(仅主程序)作为编译门禁,而非 `go build ./...`(避免上游 broken examples 阻塞);或在 `.goreleaser.yml` 显式 `dirs: ["."]` 确保只编主程序。文档标注 examples 为已知 broken。
- **回归测试建议**: CI 加 `go build .` 步骤。
- **工作量**: 15 min

#### P1-4 · Action 版本不一致(维护熵)
- **文件**: `release.yml`(checkout@v3, setup-go@v4, goreleaser-action@v4) vs `build-win7.yml`/`build.yml`(checkout@v4, setup-go@v5, goreleaser-action@v6)
- **问题**: 同一仓库 3 个 workflow 混用两代 action 版本,增加维护负担与不一致风险。
- **最小修复**: 统一到较新稳定版(checkout@v4、setup-go@v5、goreleaser-action@v6、upload-artifact@v4),并配合 P1-1 的 SHA pin。
- **工作量**: 10 min

#### P1-5 · `release.yml` 权限与 fetch-depth 过宽
- **文件**: `release.yml:9-10`(`permissions: contents: write`); `release.yml:18`(`fetch-depth: 0`)
- **问题**: `contents: write` 对 `gh release create` 是必要的最小权限(可接受),但 `fetch-depth: 0` 拉全量历史无必要(快照构建不需要),拖慢 CI 且扩大攻击面。
- **最小修复**: `fetch-depth: 1`(或去掉,默认浅克隆);权限保持 `contents: write`(release 必需)。
- **工作量**: 5 min

### P2(优化) — 6 项

#### P2-1 · testdata 私钥硬编码入库
- **文件**: `mylib/ssh/testdata/keys.go:21-148`(DSA/RSA/ECDSA/OPENSSH 私钥 PEM)
- **问题**: 测试用私钥直接写在 .go 源码。当前是上游 fixture(非真实凭证),但若误放真实密钥会同样入库。建议外置为 `testdata/*.pem` 或明确标注。
- **工作量**: 20 min(外置文件 + 改引用)

#### P2-2 · `.gitignore` 不严谨
- **文件**: `.gitignore`
- **问题**: 未忽略 `.env`、未忽略 `go.sum` 外的凭证、未绝对路径忽略 `result*`。当前未发现误提交,但防护不足。
- **最小修复**: 追加 `.env*`、`*.pem`、`*.key`、`creds/` 等。
- **工作量**: 5 min

#### P2-3 · `build-win7.yml` 冗余
- **文件**: `build-win7.yml` 整体
- **问题**: 该 workflow 注释自己承认"main release 已产出 Win7 二进制",且配了会失败的 UPX。实际是冗余 + 带 bug 的 workflow。
- **最小修复**: 要么修 P1-2 让它真可用,要么删除(主 release 已覆盖 Win7)。建议保留但修 P1-2。
- **工作量**: 含 P1-2

#### P2-4 · `release.yml` 无编译门禁
- **文件**: `release.yml` 缺少 `go build .` / `go vet .` 前置检查
- **问题**: 直接 build + 发布,若主程序编译失败,报错在 goreleaser 阶段(信息不直观)。
- **最小修复**: 在 goreleaser build 前加 `go build .` 步骤快速失败。
- **工作量**: 5 min

#### P2-5 · 缺少治理三件套
- **文件**: 仓库根
- **问题**: 无 `SECURITY.md` / `CONTRIBUTING.md` / `CODEOWNERS` / `dependabot.yml`。
- **最小修复**: 加 `SECURITY.md`(安全披露渠道)+ `dependabot.yml`(禁 major 自动更新,防供应链波动)。
- **工作量**: 30 min

#### P2-6 · 上游代码 Not assessed 标注
- **问题**: 73k LOC 上游引擎(mylib/*, Plugins/*, PocScan/*)未审计,存在潜在未发现的漏洞(安全工具本身处理不可信网络输入)。
- **建议**: 本次不做(超出范围),标注为已知盲区;后续可单独对网络输入边界做 security 审计。
- **工作量**: 不在本次闭环

## 覆盖说明
- **已审**: 本 fork 自定义改动(CI workflows、.goreleaser.* 配置、sync-upstream.sh、README/SVG、go.mod/go.sum 主程序编译)、密钥/路径静态扫描、action 版本与权限。
- **未审**: 上游 73k LOC 引擎代码(Not assessed,第三方 inherited);完整依赖树(go list 超时,未全量);运行时行为/模糊测试。
- **重要命令**: `go build .`(✅ 通过)、`go vet .`(✅ 无问题)、`grep` 密钥/路径扫描(无真实泄露)。
