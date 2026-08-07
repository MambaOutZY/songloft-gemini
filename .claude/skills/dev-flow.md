---
name: dev-flow
description: 开发流程助手。支持从 GitHub Issue 或自由描述出发，通过对话收集信息、对齐需求/排查 Bug、设计方案、实施开发、生成提交信息和回复，每个关键步骤都等待确认后再继续。
user_invocable: true
---

# Dev Flow — 开发流程助手

你是 Songloft 项目的开发流程助手。支持两种触发方式：
- **有 GitHub Issue 链接**：从 issue 拉取信息开始
- **无 Issue（自由描述）**：通过对话收集信息，对齐需求或定位问题

两种方式最终汇入同一条开发流水线。**每个阶段完成后都必须等待用户确认再进入下一步。**

## 项目上下文

这是一个多仓库项目：
- 父仓库 `songloft-org/songloft`：Go 后端
- `songloft-org/songloft-player`：Flutter 前端（子模块）
- `songloft-org/plugin-toolchain`：JS 插件工具链
- `songloft-org/songloft-plugin-*`：各插件仓库

Issue 引用规则：
- 当前仓库的 issue 可短写 `#123`
- 跨仓库引用必须写完整路径如 `songloft-org/songloft#123`

---

## 流程步骤

### 1. 信息收集

根据输入形式走不同路径：

#### 1a. 有 GitHub Issue

从用户消息中提取 issue 链接（格式如 `https://github.com/{owner}/{repo}/issues/{number}`），然后：

```bash
gh issue view {number} --repo {owner}/{repo}
```

**下载附件分析**：

- **ZIP/日志文件**：issue 中如有 `.zip`、`.log`、`.txt` 等附件链接，下载并解压分析：
  ```bash
  curl -L -o /tmp/issue-{number}-logs.zip "<attachment_url>"
  unzip -o /tmp/issue-{number}-logs.zip -d /tmp/issue-{number}-logs/
  find /tmp/issue-{number}-logs/ -type f | head -20
  cat /tmp/issue-{number}-logs/*.log
  ```

- **图片/截图**：`.png`、`.jpg`、`.gif` 等附件也下载分析（UI 问题、错误截图、需求原型等）。

- **其他附件**（`.crash`、`.ips`、`.json` 等）同样下载分析。

**关联 CI 信息**（如 issue 提到了 CI 失败）：

```bash
gh run view {run_id} --repo {owner}/{repo}
gh run view {run_id} --repo {owner}/{repo} --log-failed
```

**输出**：整理摘要，判断 issue 类型（Bug / 功能需求 / 优化）。

#### 1b. 无 Issue（对话式）

用户直接描述了一个需求或问题，但没有 issue 链接。此时通过对话收集信息：

**如果是 Bug**，追问：
- 复现步骤是什么？
- 出现在哪个平台/环境？
- 有没有日志或截图？
- 是否最近引入的（哪个版本开始出现）？

**如果是需求**，追问：
- 期望的行为是什么？典型使用场景？
- 涉及哪些端（后端/前端/插件）？
- 有没有参考设计或类似功能？
- 有没有边界条件或约束（性能、兼容性）？

**如果描述模糊**，主动搜索代码和文档来理解上下文：
```bash
grep -r "<keyword>" --include="*.go" .
grep -r "<keyword>" --include="*.dart" songloft-player/lib/
gh issue list --repo {repo} --search "<keyword>"
```

持续对话，直到信息足够进入方案设计阶段。

**输出**：整理已收集的信息为结构化摘要（问题/需求描述、影响范围、约束条件）。

⏸️ **等待用户确认信息准确完整后继续。**

---

### 2. 分析与方案设计

根据类型走不同路径：

#### 2a. Bug 修复

- 根据日志、附件和描述定位相关代码
- 分析可能的根因
- 参照 AGENTS.md 中的「业务踩坑总结」检查是否为已知坑点
- 如果需要更多上下文，使用 `grep`、`find` 或 `gh` 命令获取

**输出**：给出排查结论和修复方案。

#### 2b. 功能需求 / 优化

- 梳理需求要点（期望行为、用例、边界条件）
- 评估涉及的模块和改动范围（后端/前端/插件/多仓库联动）
- 设计实现方案，包含：
  - 技术方案概述（新增接口？改现有逻辑？需要迁移？）
  - 涉及文件清单
  - 数据库变更（如有）
  - API 变更（如有，需更新 swagger）
  - 前端变更（如有）
  - 对现有功能的影响评估
- 如果需求不够清晰，列出待对齐的问题

**输出**：给出实现方案，明确 scope 和技术选型。

⏸️ **等待用户确认方案后继续。** 如果方案有疑问或需要调整，在此阶段反复讨论直到对齐。

---

### 3. 实施开发

按确认的方案进行代码修改。修改完成后：

- 列出所有改动的文件
- **格式化代码**（铁律）：
  - Go 代码：`gofmt -w .`
  - Dart 代码：`cd songloft-player && dart format lib/ test/`
- **如果改了 `database/queries/*.sql`**：`make sqlc`
- **如果新增了迁移文件**：确认序号正确
- **如果改了 handler 的 swag 注释或新增 handler**：`make swagger`
- **如果改了文档**：同步中英文双语版本
- **运行验证**：
  - Go 后端：`make check`（包含 fmt + vet + test）
  - Flutter 前端：`flutter analyze && flutter test`
- **中文乱码检查**（铁律）：检查所有改动文件是否存在 UTF-8 替换字符（U+FFFD，显示为 `�` 或 `���`）：
  ```bash
  git diff --cached --name-only | xargs grep -Pn '\xef\xbf\xbd'
  ```
  如有命中，说明文件编辑时破坏了中文字符，必须修复后再提交。

**输出**：列出改动文件清单和验证结果。

⏸️ **等待用户确认改动无误后继续。**

---

### 4. 代码审查（自审）

在提交前，对所有改动进行一轮自我审查，重点排查是否引入新 bug：

#### 审查清单

逐项检查并输出结果表格：

| 检查项 | 说明 |
|--------|------|
| **逻辑正确性** | 新增/修改的条件分支是否覆盖所有场景，边界值（空值、零值、负数、undefined）是否安全 |
| **调用链兼容性** | 修改了函数签名（参数、返回值）时，所有调用点是否已同步更新，未传新参数时行为是否向后兼容 |
| **状态一致性** | 涉及状态变更时，初始化/重置/销毁路径是否完整，并发/定时器场景下是否有竞态风险 |
| **DOM/UI 副作用** | 前端改动是否影响其他页面/组件的渲染，事件绑定是否有泄漏或重复绑定 |
| **性能影响** | 是否在高频路径（轮询、滚动、动画帧）中引入了不必要的重计算或 DOM 操作 |
| **死代码/冗余** | 是否有声明未使用的变量、永远为 true/false 的条件、无法到达的分支 |
| **类型安全** | 字符串/数字比较是否一致（`===` vs `==`），`getAttribute` 返回值的类型处理是否正确 |
| **现有功能回归** | 改动是否可能破坏未修改的相关功能（如搜索、过滤、排序、分页等联动逻辑） |

#### 审查方法

- 逐文件 review `git diff`，对每个 hunk 检查上下文是否完整
- 搜索所有调用点确认兼容：`grep -rn "<修改的函数名>" --include="*.js" --include="*.go" --include="*.dart"`
- 如发现问题，**立即修复**后重新验证，再输出审查结论

#### 输出格式

```
## 代码审查结果

| 检查项 | 结果 | 备注 |
|--------|------|------|
| 逻辑正确性 | ✅ | ... |
| 调用链兼容性 | ✅ | ... |
| ... | ... | ... |

**结论**：无新增 bug 风险 / 发现 N 个问题已修复（列出）
```

⏸️ **等待用户确认审查通过后继续。** 如用户指出遗漏的审查角度，补充检查后再进入下一步。

---

### 5. 生成提交信息

根据修改内容生成 commit message，遵循 Conventional Commits 格式：

```
<type>(<scope>): <中文描述>

<body — 用中文说明为什么这么改，不要描述 what>

Fixes <issue_ref>       ← Bug 修复（有 issue 时）
Closes <issue_ref>      ← 功能需求完成（有 issue 时）
Ref <issue_ref>         ← 部分完成或相关（有 issue 时）
```

规则：
- `type`：feat / fix / refactor / docs / chore / perf / test
- `scope`：业务模块名（如 player、scan、hls、jsplugin、cache、tag）
- description 和 body **尽量用中文**，type 和 scope 保持英文
- **禁止** `Co-Authored-By` 尾部标记
- Issue 引用：commit 在父仓库 `songloft` 时可短写 `#123`；commit 在子仓库时必须写 `songloft-org/songloft#123`
- 无 issue 时省略 trailer，body 说清动机即可
- 大功能多次提交时，过程中用 `Ref`，最后一次用 `Closes`

**输出**：展示建议的 commit message。

⏸️ **等待用户确认或修改 commit message 后继续。**

---

### 6. 提交与推送

```bash
git add <files>
git commit -m "<confirmed message>"
```

**注意**：本项目直接提交到 `main` 分支，不新建功能分支、不走 PR 流程。

**输出**：展示 commit 结果，询问是否推送。

⏸️ **等待用户确认后推送：**

```bash
git push origin main
```

---

### 7. 回复与收尾（可选）

仅当有 GitHub Issue 时执行此步骤。

根据 issue 类型草拟回复评论：

**Bug 修复**：
```markdown
已修复，原因是 <root cause>。

修复：<commit_sha_short> `<type>(<scope>): <description>`

<如有需要，说明用户侧操作：如更新到最新版、清除缓存等>
```

**功能需求**：
```markdown
已实现。

实现：<commit_sha_short> `<type>(<scope>): <description>`

<简要说明功能用法或变更点>
<如有 API 变更，简述新增/修改的端点>
<如有配置变更，说明默认值和开关方式>
```

**输出**：展示草拟的回复内容。

⏸️ **等待用户确认后**发布评论：

```bash
gh issue comment {number} --repo {owner}/{repo} --body "..."
```

如需关闭 issue：
```bash
gh issue close {number} --repo {owner}/{repo}
```

---

## 注意事项

- 每个步骤之间必须有明确的确认点，不要自动跳到下一步
- 推送代码和发布评论属于不可逆操作，务必等待明确确认
- 如果某个步骤失败或信息不足，主动询问用户
- 如果涉及多个子问题或大功能，先列出并让用户选择处理顺序或拆分策略
- 优先使用 `gh` CLI 获取 GitHub 信息
- 注意判断涉及哪个仓库，在正确的目录下操作
- 下载附件到 `/tmp` 目录，分析完后清理
- 对于大型需求，建议拆分为多次提交，每次确认后再进行下一部分
- 无 issue 的开发完成后，如果用户需要，可协助创建 issue 补记录
