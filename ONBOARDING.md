# Loop Engine 接入指南

把这套自治开发循环搬到一台**只有内网模型、只有工蜂/TAPD** 的机器上,并接到一个真实
仓库上。读者是那台机器的运维者;本文只给步骤与判据。

一句话形态:在一个 `LOOP_STAGE=orchestrate claude` 会话里发一句 `/loop 1h /tick`,外层
agent 每隔一段时间用 `/tick` 启动一次 `$HOME/.local/bin/loop-runner tick`——P0 复位
(`loopctl tick prepare`)
→ 侦察 → 纯代码裁决 → 跑一条管线 → 门禁 → 开 MR 或升级给人 → 收尾
(`loopctl tick close`)。所有机械判断在 `loopctl`(一个 Go 单二进制)里,模型只做模型
该做的事。

**两条硬前置,忘了任何一条的代价都不是报错**:

- 会话必须带 `LOOP_STAGE=orchestrate`。不带的话 `scripts/hook.sh` 的"没有 stage 就放行"
  捷径会让**这一整天所有角色的写**绕过写执法层,而表现是一切照常。`loopctl tick prepare`
  第 ① 步因此会直接 `exit 3` 让 tick 起不来——一个当场看得见的代价,换掉了那个看不见的。
- `config.budget.daily_tokens` 必须按本机的模型串配齐。没配上限的模型 = 不受限的模型
  (`budget check` 会 `exit 2` 停跑,不给默认值)。

---

## 0. 目录

1. [前置条件](#1-前置条件)
2. [安装与 `loopctl init`](#2-安装与-loopctl-init)
3. [`config.yaml` 逐项说明(含必改项)](#3-configyaml-逐项说明)
4. [接平台:MCP 对接清单与 bot 最小权限](#4-接平台mcp-对接清单与-bot-最小权限)
5. [hooks 与沙箱](#5-hooks-与沙箱)
6. [接入顺序](#6-接入顺序)
7. [故障排查](#7-故障排查)

---

## 1. 前置条件

| 项            | 要求                     | 怎么确认                              |
| ------------- | ------------------------ | ------------------------------------- |
| Go            | 1.22 以上                | `go version`                          |
| git           | 2.20 以上(要 worktree)  | `git --version`                        |
| Node.js       | 22.18 以上(运行 Agent SDK runner) | `node --version`             |
| Claude Code   | 2.1.220 以上(由 Agent SDK 启动) | `claude --version`           |
| mutation 工具 | `gremlins`(默认)或 `go-mutesting` | 见下                        |
| bash          | 4 以上                   | `bash --version`                      |
| 目标仓库      | 一个 Go **module** 仓库,有远端 origin | `go list ./...` 能列出包 |

**mutation 工具的版本必须钉住**。`config.mutation_tool_version` 与工具自报版本不符即
`exit 3`,这不是洁癖:变异体的采样身份是 `file:line:col:type + 同位序号`,mutator 命名或
column 口径一变,"同种子复跑得同集合"就静默失效,而没有任何门禁看得见。安装后先确认:

```bash
gremlins --version        # 把输出里的版本号填进 config.mutation_tool_version
loopctl mutate --wt <某个 worktree> --scope diff   # 版本不符会直接告诉你 expected/actual
```

**网络白名单**。这台机器要能访问:

- 内网模型端点(Claude Code 配置的那个);
- 工蜂 / TAPD 的 MCP server;
- 目标仓库的 git 远端。

**不需要**、且应当**拒绝**的:公网。测试与变异测试**在断网沙箱里跑**(见第 5 节),
仅放行 loopback。

---

## 2. 安装与 `loopctl init`

```bash
# 1. 在已经 clone 的 loop-engr 源码仓里一次性构建并安装
./scripts/install.sh
$HOME/.local/bin/loopctl version  # {"version":"v2"} —— 记住这个值

# 2. 在目标仓库里初始化
cd /path/to/your-repo
$HOME/.local/bin/loopctl init
```

安装默认落到 `$HOME/.local/bin`,不需要 sudo、GitHub Packages 或 `gh login`。如果机器约定
了别的用户级前缀,安装和运行时都设置同一个 `LOOP_INSTALL_PREFIX`。把该目录加入 `PATH`
后也可以直接使用 `loopctl` 和 `loop-runner`。

`init` 写两类文件,区别是一条执法边界而不是习惯:

- **内嵌真相源(护栏)**:`.loop/schemas/`、`.loop/prompts/`、`.loop/policy/rubric.md`、
  `.loop/PLATFORM.md`、`scripts/hook.sh`。它们**随二进制发版**,
  `loopctl config validate` 逐字节核对,改了就 `exit 3`。要改就改 loopctl 再发一版。
- **你自己的**:`.loop/config.yaml`、`.loop/policy/blocklist.yaml`、
  `.loop/policy/test-deps.yaml`、`.loop/ratchet.yaml`、`.loop/lessons/`、
  `.claude/skills/tick/SKILL.md`、`.claude/settings.json`、本文件。`init` 只给一份起始版本,
  **此后永不覆盖**——包括 `--force`。`settings.json` 是个例外的一半:文件是你的,但里面
  那条 PreToolUse 接线会被 `config validate` 核(第 5 节)。

重跑 `init` 是安全的:内容一致的记 `unchanged`,不写盘。护栏被改过时它**拒绝**覆盖并
`exit 1`,要覆盖得显式 `loopctl init --force`。

**`init` 之后 `config validate` 会失败,这是设计**:它报出的每一条都是还没按本仓填的
东西。下一节就是同一份清单的人类可读版。

---

## 3. `config.yaml` 逐项说明

改完每一项后跑 `loopctl config validate`,直到 `errors` 为空。

### 3.1 必须按本仓调整的(⚠)

| 键                          | 改成什么                                                                 |
| --------------------------- | ------------------------------------------------------------------------ |
| `models.*`                  | **换成内网可用的模型名**。九个角色分三档:cheap(scout/secretary/librarian)、mid(planner/ut_writer/ut_reviewer/spec_extractor)、top(implementer/arbiter)。档位错配的表现是"贵而不准"或"便宜到读不懂 diff",不是报错。 |
| `default_branch`            | 本仓默认分支(`master` / `main` / …)。基准是 `origin/<它>`,不是本地 HEAD。 |
| `coverage.repo_target`      | 按本仓**现状**定,不是理想值。先跑影子运行看分布,把目标定在略高于现状处;够不到的目标等于没有目标(early exit 永不触发,每个 tick 都会去写测试)。 |
| `coverage.pkg_floor`        | 同上。低于它的包在 direct_l2 里整体优先。                                 |
| `coverage.min_contract_density` | 契约密度下限。低于它、且该包没有既有外部测试时,不写测试而是开文档债票。影子运行会给出本仓的 doc_ratio 分布,按它定。**定得太低的代价是特征化测试**:把当前行为连 bug 一起冻结成断言,然后 ratchet 把这条线锁死。 |
| `mutation_tool_version`     | 本机已安装工具的自报版本,逐字。                                          |
| `loopctl_version`           | `loopctl version` 的输出。二进制升级时**两处一起改**,否则 `exit 3`。      |
| `paths.wt_root`             | worktree 根目录,**必须在仓库之外**(剥体树不能被 `go build ./...` 看见)。 |
| `platform.mode`             | 接真实平台就是 `mcp`。                                                     |
| `platform.mcp.servers/tools` | 见第 4 节；连接显式配置，不继承个人 Claude settings。                       |
| `exclude`                   | 生成物与 vendor。影子运行会告诉你本仓是不是 vendor 模式。                  |
| `budget.daily_tokens`       | **键是 Agent SDK `modelUsage` 返回的完整 model 串**(如 `claude-haiku-4-5-20251001`),不是 `models:` 别名。值是当日(UTC+8)`input + cache_creation + output` 的上限(不含 `cache_read`)。初值是占位数,按第 6 节第 5 步用真跑数据校准。**当日出现没有预算键的实际模型会 `exit 2` 停跑**。 |
| `budget.daily_usd`          | 可选。只有确认网关返回的 `costUSD` 与公司计费一致后再按完整 model 串配置；配置后 cost 缺失会 fail-closed，token 或 USD 任一触顶即停。 |

### 3.2 按团队习惯调的

| 键                              | 含义                                                              |
| ------------------------------- | ----------------------------------------------------------------- |
| `limits.soft_files` / `soft_loc` | 软限。超了就发拆分建议,不写代码。                                 |
| `limits.hard_factor`            | 硬限 = 软限 × 它。软硬之间且计划完成度够就允许收尾。               |
| `limits.plan_completion_min`    | 上句里的"够"。完成度 = 计划文件里被真正改到的比例。                |
| `limits.tick_wall_min` / `role_wall_min` | 整个 tick 与单次角色调用的时间预算。CI 不在 tick 内等待，由下一轮读取。 |
| `limits.role_max_turns.<role>`  | 可选的快速工具循环熔断。只限制列出的角色；未列出的角色仍受 `role_wall_min` 与每日预算约束。真实模型/平台的工具调用节奏不同，应按运行数据校准，不是 runner 常量。 |
| `retries.l3` / `l2_writer`      | 重试预算。**随 ticket 持久化,不随 tick 归零**——这是防"失败票变成跨 tick 慢速烧钱永动机"的那道闸。 |
| `quality.*`                     | 门禁阈值:diff 行覆盖、两条管线各自的 mutation 分数线、每轮取几条行为、等价 mutant 允许比例、变异体上限。 |
| `ttl.claim_hours`               | claim 多久算过期(过期后 P0' 会把自己的那把释放掉)。v1 只有这一条:提问与 MR 的催办没有执行者,连同它们的 TTL 一起继续后移。 |
| `quota.wip_mrs`                 | 同时在飞的新活数。初值 1:先把一件事做完。                          |
| `quota.agent_tickets_per_day`   | loop **自己开票**的日上限(可测性债 / 文档债 / 覆盖率票 / 主干红票)。这是防刷屏的闸,调大之前先看看人类受不受得了。 |
| `quota.grill_questions` / `grill_rounds` | 单张票的提问总数与轮数上限。                              |
| `quota.lessons_cap`             | lessons 库总条数上限。                                             |
| `trusted_responders`            | 叠加在平台角色校验之上的可信作者名单。**只有可信作者的 `@loop` 命令算数**。 |
| `labels_prefix`                 | 所有 label 的统一前缀。改它要同时确认平台上没有同名遗留 label。     |
| `docs_autofix`                  | **保持 `false`**。`true` 的那一半(自动补注释走 L3 人审)没实现,写 `true` 会被 `config validate` 直接拒掉。 |

---

## 4. 接平台:MCP 对接清单与 bot 最小权限

### 4.1 对接清单

契约在 **`.loop/PLATFORM.md`**——十个平台动作,每个写明要什么能力、输入输出哪些字段、
落到 `WorldReport` 或 `ActionResult` 的哪一格。先读它,再填绑定:

```yaml
platform:
  mode: mcp
  mcp:
    servers:
      gongfeng:
        transport: http
        url_env: LOOP_MCP_GONGFENG_URL
        headers_env: {Authorization: LOOP_MCP_GONGFENG_AUTH}
      tapd:
        transport: stdio
        command: /company/bin/tapd-mcp
        args: []
        env_from: [TAPD_TOKEN]
    tools:
      read_ticket: {server: tapd, tool: <工具名>} # 读 ticket:按 id 取、按 label 列
      read_mr: {server: gongfeng, tool: <工具名>}
      read_ci: {server: gongfeng, tool: <工具名>}
      read_comments: {server: gongfeng, tool: <工具名>} # 含作者角色
      add_label: {server: gongfeng, tool: <工具名>}
      remove_label: {server: gongfeng, tool: <工具名>}
      comment: {server: gongfeng, tool: <工具名>}
      open_mr: {server: gongfeng, tool: <工具名>}
      open_ticket: {server: tapd, tool: <工具名>}
```

runner 固定 `settingSources: []`、`strictMcpConfig: true`。URL/token 可通过上面的环境变量
引用提供；token 不得写进 config。缺变量或连接失败时 fail-closed。

三件容易漏的:

1. **作者角色判定是 `read_comments` 的一半**,不是附加项。"回答有效者 = reporter |
   assignee | maintainer ∪ `trusted_responders`";判不出角色,就判不出一条回答算不算数、
   一条 `@loop stop` 是不是可信作者发的。工具给不出角色时,要么另配一条能查仓库成员的
   能力,要么把人写进 `trusted_responders`——**不得默认可信**。
2. **`open_mr` / `open_ticket` 必须能查重**:前者按 `(ticket, source_branch)`,后者按
   `(ticket_type, title)`。"更新既有 MR"与"主干红了只开一张票"这两件事都靠它。
3. **建分支不在 `tools` 里**:它是 Implementer 在自己 worktree 里的 `git push`,不是
   MCP 动作。但 bot 权限要按十个动作配,别漏。

绑定缺口不用自己找:跑一次影子运行,`SHADOW-REPORT.md` 第 6 节会列出来。

### 4.2 bot 账号最小权限

| 能力                          | 给不给 | 为什么                                                     |
| ----------------------------- | ------ | ---------------------------------------------------------- |
| 读 ticket / MR / CI / 评论    | ✅     | Scout 的全部工作                                            |
| 评论 ticket / MR              | ✅     | 提问、失败摘要、计数镜像、拆分建议                          |
| 打 / 摘 label                 | ✅     | claim、失败标、needs-clarification                          |
| 推送 `loop/*` 分支            | ✅     | WIP 保全与开 MR 前的那一次推送                              |
| 建 MR                         | ✅     | 门禁绿之后                                                  |
| 建 ticket                     | ✅     | 可测性债 / 文档债 / 覆盖率票 / 主干红票(受日配额)          |
| **合并 MR**                   | ❌     | 人审是整条链路的出口,自动合并等于把它删掉                  |
| **关闭 ticket**               | ❌     | 关票是人的判断;loop 只会打标与评论                          |
| **推送非 `loop/*` 分支**      | ❌     | 尤其是默认分支                                              |
| **force-push**                | ❌     | 永久禁止                                                    |
| **改仓库设置 / CI 配置 / 分支保护** | ❌ | I10:护栏不得自改                                          |

**分支保护要在平台侧真的配上**:默认分支受保护、只允许 bot 推 `loop/*`、拒非快进推送。
prompt 里那几行"不许 force-push"只是说明书——执法在这里。

---

## 5. hooks 与沙箱

### 5.1 PreToolUse hooks(写执法层)

写执法层按**目标路径**拒写的规则:

```
stage=l2 : 仅允许写 **/*_test.go 与 **/testdata/**;
           go.mod / go.sum 按 .loop/policy/test-deps.yaml 白名单判
stage=l3 : 拒 .loop/policy/blocklist.yaml 里的全部路径;拒 **/*_test.go 与 **/testdata/**
任意 stage: 拒 .git/**、CI 配置、I10 自改清单
```

**这次写属于哪一行,由写入路径推断**:

| 写到哪                                      | 按哪一行判  | 为什么是这个路径 |
| ------------------------------------------- | ----------- | ---------------- |
| 剥体树(带 `.stripped-by-loopctl` 标记的目录) | `stage=l2`  | UT-Writer 只在剥体树里写测试 |
| `paths.wt_root` 下的实现 worktree            | `stage=l3`  | Implementer 只在自己的 worktree 里改代码 |
| 其余(主仓工作树、wt_root 之外)              | 只判"任意 stage" | 没有任何管线往这里写;I10 与 CI 配置那一行照常拦 |

`LOOP_STAGE` 只剩两个作用:**未设**即直接放行(那不是 loop 在跑,是你在改自己的仓库);
设成 `l2`/`l3` 可以**覆盖**路径推断,给下面那两条手工验证用。真 tick 跑在
`LOOP_STAGE=orchestrate` 下,阶段全部由路径推断出来。

**`init` 已经把它装好了**,三层各自只做一件事:

| 层                       | 做什么                                                            |
| ------------------------ | ----------------------------------------------------------------- |
| `.claude/settings.json`  | 接线:`PreToolUse` 匹配 `Edit`/`Write`/`MultiEdit`/`NotebookEdit`/`Bash`,`command` 指向 `scripts/hook.sh`。**这份文件是你的**——permissions、env、别的 hook 随便加,只别删掉这一条。 |
| `scripts/hook.sh`        | 转发:stdin 原样喂给 `loopctl hook check`。**fail-closed**——loopctl 找不到、崩了、没吐出裁决,一律当拒绝。唯一例外是 `LOOP_STAGE` 未设(那不是 loop 在跑,是你在改自己的仓库),直接放行。 |
| `loopctl hook check`     | 判定:全部规则。判据是上面两份 policy 文件。**恒 exit 0**,拒绝写在 JSON 里——hook 协议里非零退出是"钩子自己坏了",与"这次写被拒"不是一回事。 |

`loopctl config validate` 会核这条接线在不在(不在即 `exit 3`)。它核的是**接线**不是整份
文件的 hash,理由与 blocklist 的 I10 子集核对一样:必须在的那部分逐条核,其余自由。

自己验一次(不需要跑 tick):

```bash
echo '{"tool_name":"Write","cwd":"'"$PWD"'","tool_input":{"file_path":"'"$PWD"'/.loop/config.yaml","content":"x"}}' \
  | LOOP_STAGE=l3 ./scripts/hook.sh          # 应当 deny,理由指名 I10
echo '{"tool_name":"Write","cwd":"'"$PWD"'","tool_input":{"file_path":"'"$PWD"'/foo_test.go","content":"x"}}' \
  | LOOP_STAGE=l3 ./scripts/hook.sh          # 应当 deny(测试是 L2 的产出)
```

`git push` 是 `stage=l3` 下**允许**的命令(它不写仓内路径);"只能推 `loop/*`、不许
force-push"由平台侧执法(第 4.2 节),不由 hook 执法。

三行现在**全部在执法**。
所以"Implementer 不写测试"与"L2 只写测试文件"不再只有 prompt 在管——
前提是这两类活确实各自跑在自己那棵树里,而那两棵树都是 loopctl 建的(`wt create` /
`strip --out`),不由任何 prompt 决定。

### 5.2 沙箱

`loopctl test run` 与 `loopctl mutate` 必须在**断网**环境里执行,且**放行 loopback**
(`httptest` 依赖它),不注入任何 secret。做法取决于这台机器:

- Linux + systemd:给 tick 的 service 加 `PrivateNetwork=yes`,并保留 `lo`;
- 容器:`--network=none` 起测试容器,或 network namespace 里只留 `lo`;
- 退而求其次:iptables/nftables 出站默认拒绝,放行 `127.0.0.0/8` 与 `::1`。

网络隔离的执法主体是沙箱。`loopctl lint tests` 只拦 `import "net"` 这个拨号入口,
`net/http`、`httptest`、`netip` 一律放行——它是提醒,不是围墙。

`.gitignore` 里要有 `.loop/runtime/`(`init` 会放一份 `.loop/.gitignore`)。它是可丢弃
缓存:丢了最坏的结果是重新规划一次,计数在平台上,WIP 在已推送的分支上。

---

## 6. 接入顺序

一步一步来,每步都有明确的通过判据。**不要跳步**——后面每一步的诊断成本都比前一步高
一个量级。

### 第 1 步:影子运行

```bash
loopctl config validate          # errors 必须为空

LOOP_STAGE=orchestrate LOOP_SHADOW=1 claude
> /tick                          # 只发这一次,不要 /loop
# 会话退出后:
loopctl shadow report --world .loop/runtime/world.json \
  --density .loop/runtime/density.json --coverage .loop/runtime/coverage.json
#                                  读 .loop/runtime/SHADOW-REPORT.md
```

它只跑 P0 + Scout + 裁决,**一个字节都不写**:Secretary 起不来、loopctl 拒掉自己所有
写命令(含 `tick close`)、任何管线都不进入。`LOOP_STAGE` 照样要设:P0 第 ① 步不分
影子与否——一条无条件的规则,比一条"影子下例外"的规则少一个可以走错的分支。

通过判据:报告出得来,第 1 节的裁决是你**看得懂并且认同**的动作,第 7 节的缺口清单
你能逐条解释。

### 第 2 步:校准三条债

影子报告直接对应前两条:

1. **vendor 剥体策略**:本仓是 vendor 模式吗?`vendor/**` 在 `config.exclude` 里吗?
   不在的话,剥体树会把整个 vendor 目录搬一遍,diffstat 会把它算成 loop 改的。
2. **工蜂 CI 路径**:报告第 7 节列出"存在但没被 blocklist 拦住"的 CI 配置路径。逐条加进
   `.loop/policy/blocklist.yaml`(I10 那几条只能加不能减)。**多拦无害,漏拦即破口**。
3. **CI 结果读取**:`read_ci` 拿回来的 `status` 要能映射成 green / red / pending。跑
   `loopctl decide --world .loop/runtime/world.json` 对着平台看一眼:`main_green` 与
   `my_mrs[].ci` 是不是与平台上显示的一致?默认分支若被平台**明确判定为未配置 CI /
   从未有流水线**(包括平台文档明确规定为该语义的 404),`main_green=true`;没有失败的
   CI 不能叫 `main_red`。权限、路由、工具故障或含义不明的 404 仍填保守值
   (`main_green=false`);MR 取不到结论仍为 `ci=pending`。

改完再跑一次影子运行,直到缺口清单空掉、或剩下的每条你都能说清为什么留着。

### 第 3 步:单仓 direct_l2 试跑

**前置(两条,都必须先做完)**:

1. **hooks 已装好**:`loopctl config validate` 通过即说明接线在(第 5.1 节)。
2. **目标仓 CI 里有 `loopctl ratchet check`**。direct_l2 的产物之一就是被
   `ratchet raise` 抬高的 per-pkg floor,而"抬上去之后没人再往下掉"这件事**只由目标仓
   自己的 CI 执法**——loop 不跑你的 CI,也不会去改它(那是 I10 blocklist 里的路径)。
   缺了这一条,第 3 步看起来会完全正常:floor 照抬、MR 照合、覆盖率照掉,没有任何一层
   会说话。

   加法是在每次流水线里加一步(退出码即判定,`1` = 有包掉到 floor 以下):

   ```yaml
   # 举例:GitHub Actions 的一个 step;别的 CI 同理,重点是这条命令真的跑
   - name: ratchet
     run: loopctl ratchet check
   ```

   加完再跑一次影子运行(第 1 步那套命令):报告第 7 节里那条"CI 配置里找不到
   `ratchet check`"应当消失。**没消失就是没接上**——报告找的是仓内已存在的 CI 配置
   文件里出现过 `ratchet check` 这几个字,通过 make 目标或复合 action 间接调用时,
   把那句话原样写进注释或脚本名,让它看得见。

先只开这一条管线:它只加测试、不改实现,是风险最低的写路径。做法是让裁决走不到别的
级别——最简单的是暂时不给候选 ticket 打进候选集(平台侧不派活给 bot)。

```bash
LOOP_STAGE=orchestrate claude
> /tick                          # 一次真 tick,先不要 /loop
```

通过判据:开出一张覆盖率票、一个只含 `*_test.go` 的 MR、ratchet 抬高的 floor 在 MR 里;
人工评审这个 MR——**这是你第一次真正验"它写的测试值不值得要"**。

跑三五个 tick,看 `.loop/runtime/STATE.md` 与平台上的票。此时 `min_contract_density`
定得对不对会明显暴露:如果 MR 里的测试大多在断言"当前实现就是这样",那就是密度定低了。

### 第 4 步:开 implement 管线

给一张真正的需求票,确认它被 grill(按 rubric 六格提问)、被回答、被规划、被实现、
过门禁、开 MR。前几张票**盯着看**,尤其是:

- Planner 的拆分建议是不是合理(软限定得对不对);
- Arbiter 的路由是不是在 l2/l3 之间摆动(摆动会自动 escalate,但频繁摆动说明反馈质量差);
- escalate 出来的失败摘要,人类**读得懂吗**——这是整条链路唯一的人机接口。

最后再把它交给 `/loop`:

```
LOOP_STAGE=orchestrate claude
> /loop 1h /tick
```

外层 agent 的职责恰好三件,一件不多:①按间隔唤起 `/tick`;②读回 TickResult 的 `budget`
与 runner 的失败信号;③**当日预算触顶时停到次日 UTC+8 00:00**。它不判优先级、不判
相位、不数 token、不在 tick 之间改任何文件——每发一次就是 `/tick` 四个字符,不多不少。

它是**省事层不是执法层**:就算它完全无视预算、每分钟唤起一次,每一次也都会在
`tick prepare` 第 ④ 步停住。

跑起来之后没人能交互确认,所以 allowlist 与 hooks 必须在这之前配完备。

公司封装的 Claude Code executable 只负责模型网关；把绝对路径显式交给 runner：

```bash
LOOP_STAGE=orchestrate \
LOOP_CLAUDE_PATH=/absolute/path/to/tclaude \
"${LOOP_INSTALL_PREFIX:-$HOME/.local}/bin/loop-runner" tick
```

这个直接入口与 `/tick` 完全同一条执行路径，适合 Desktop task、CI scheduled job 或
cron，不会产生 `/loop` launcher 的 token。定时器应防止重叠，例如 cron 可用：

排查模型或网关问题时，可临时加 `LOOP_DEBUG_AGENT_STREAM=1`。它把每个角色的模型文本、
API retry、工具名（不含参数）、SDK stderr 和最终 result 元数据实时写到 stderr；stdout
仍然只有最终 `TickResult` JSON。模型文本可能包含代码或平台内容，不要默认放进长期 cron
日志：

```bash
LOOP_DEBUG_AGENT_STREAM=1 \
LOOP_STAGE=orchestrate \
LOOP_CLAUDE_PATH=/absolute/path/to/tclaude \
"${LOOP_INSTALL_PREFIX:-$HOME/.local}/bin/loop-runner" tick
```

```cron
0 * * * * cd /path/to/repo && flock -n .git/loop-runner.lock env LOOP_STAGE=orchestrate LOOP_CLAUDE_PATH=/absolute/path/to/tclaude "$HOME/.local/bin/loop-runner" tick >>.loop-runner.log 2>&1
```

调度器解析 stdout 的 `TickResult`：`budget_exhausted` 停到次日 UTC+8；
`p0_failed` 报警；其余结果按下个调度周期继续。网关认证环境照常继承，但平台 MCP
只来自 `platform.mcp` 的显式配置；runner 固定不读取 user/project/local settings。

### 第 5 步:校准预算

`config.budget.daily_tokens` 的初值是占位数。runner 直接记录 Agent SDK 返回的每个实际
模型及 usage；网关返回可靠 `costUSD` 时也一并存入机器账本。

```bash
loopctl budget report            # 当日按模型 × 角色的账
```

跑几天,看真实分布,再把上限设在"一天正常跑完该跑的活"之上、"跑疯了"之下。

两个自查点:

- `sdk_events_counted` 突然翻番 → 检查角色是否被重复唤起。
- `sources[]` 应指向机器级账本；删除 `.loop/runtime/` 不应让当天数字归零。

---

## 7. 故障排查

### `exit 3`(环境错误:重跑没用,得修环境)

| 报错里出现                            | 成因与处置                                                                 |
| ------------------------------------- | -------------------------------------------------------------------------- |
| `loopctl_version: ... but this binary is ...` | 二进制与仓内护栏不是同一版。装对版本,或 `loopctl init --force` 重装护栏并改 config。 |
| `shipped guardrails ... drifted from the shipped copy` | 有人改了内嵌真相源(prompt / schema / rubric / PLATFORM.md / hook.sh)。确认那些改动不该保留后 `loopctl init --force`;确实需要改,就去改 loopctl 再发一版。 |
| `shipped guardrails ... not installed` | 没跑过 `loopctl init`,或文件被删。跑 `loopctl init`。                        |
| `blocklist: the I10 entries ... cannot be weakened` | blocklist 少了 I10 条目、或 `stages` 不含 `any`。补回去(只加不减)。 |
| `schemas directory missing` / `required schemas missing` | `.loop/schemas/` 没装上。`loopctl init`。                |
| `hooks: ... has no PreToolUse entry calling scripts/hook.sh` | 写执法层没接上。`loopctl init` 装一份 settings.json,或把那条 `PreToolUse` 加回你自己的那份。 |
| `mutation tool ... expected/actual`   | 工具版本与 `mutation_tool_version` 不符,或工具没装。装对版本或改 config——**不要**把这条当质量失败去重试。 |
| `canary produced 0 mutants`           | 变异工具或其报告格式已死(常见于升级后字段改名)。修工具,别调阈值。         |
| `world.json` 缺字段                    | Scout 没填全 `WorldReport`。对着 `.loop/PLATFORM.md` 查是哪个动作没绑定或读失败。 |
| `cannot resolve origin/<branch>`      | `default_branch` 写错,或没 `git fetch` 到。                                 |
| `LOOP_STAGE is not set` / `LOOP_STAGE is "..." not "orchestrate"` | 会话没带 stage。用 `LOOP_STAGE=orchestrate claude` 重开会话。**不要**把它写进 `.claude/settings.json` 的 `env`:那会让你在这个仓库里的每一次手改都掉进 hook,而"人类改护栏走人审"指的正是手改。 |
| `usage ledger: git remote origin is required` | 机器账本用规范化 origin 生成 repo-id；先配置 `remote.origin.url`。 |

### `exit 4`(当日预算触顶:今天到此为止,不是坏了)

只有 `budget check` 与 `tick prepare` 会给这个码。它与 `exit 3` 分开,是因为**处置不同**:
环境错误是"修了再跑",预算触顶是"今天别再跑了"。外层 `/loop` agent 据此停到次日
UTC+8 00:00。

想确认花在哪了:`loopctl budget report`,看 `models[]` 与 `by_role[]`。真觉得上限定低了
就改 `config.budget.daily_tokens`——但先看一眼 `by_role[]`,常见的情况是某个角色在反复
重试,而那是别的问题。

### `exit 1`(判定不过:东西在,结论是不行)

- **`config validate` 的 `errors[]`**:逐条读,每一条都是一个还没填的键。`platform.mcp.tools.*`
  未绑定是最常见的一条。
- **门禁 `missing_evidence`**:门禁**证据缺失即失败**,不给默认分。gap 里的 `artifact`
  字段指名道姓:`mutate.json` 表示整份报告没写出来(多半是上游命令 exit 2/3 挂了,去看
  那一步的 stderr);`cover_func.json#focus.increment` 表示报告在、但那个字段是 null
  (focus 指定的包不在本次结果里)。**不要**通过放宽阈值来"修"它——它说的是没测,不是没达标。
- **canary 失败 / mutation 子门禁**:`no_viable=true` 且 canary 出了 mutant,是合法空集
  (这个 scope 确实没有可变异的算子),门禁记 `not_applicable[]` 而不是满分;canary 出 0
  才是工具死了(见上,那是 exit 3)。
- **`quota` 拒绝**:`reason` 会指名是哪一条配额。日配额跨自然日归零;grill 的两条按 ticket
  累计,清不掉——那是设计,不是 bug。

### 循环行为异常

| 现象                                   | 先看哪里                                                              |
| -------------------------------------- | ----------------------------------------------------------------------- |
| 每个 tick 都 `early_exit(coverage_unknown)` | P0 的覆盖率缓存没建起来。手工跑 `loopctl cover func --wt <base-wt>` 看它为什么失败。 |
| 每个 tick 都在同一张票上重试到 escalate | 摆动检测本应更早停住;看 `reports/gate.json` 的 `signature` 与 Arbiter 的 Route。反馈质量问题,不是重试次数问题。 |
| 一张票再也不动了                        | 它带着 `<prefix>attempted-failed:*` 标。可信人员可在原 ticket/MR 评论 `@loop retry`，由 Secretary 精确摘标；也可人工摘标。两种方式都保留历史重试计数，不要另开重复票。 |
| 某个包再也不被选中                      | 同上,它的覆盖率票带着失败标。                                            |
| 主干红着,什么都不干                    | 这是对的:红主干屏蔽全部写管线(1/2/4/6),只留沟通(3/5)。修主干。       |
| 分支推了但没有 MR                       | 崩在 push 与开 MR 之间。下一个 tick 的优先级 1.5 会补开——前提是那个 ticket 的 `reports/gate.json` 还在(`runtime/` 被删就读不到,于是保守判为没绿,交给人)。 |
| `@loop retry` 没反应                    | 确认评论者属于 reporter、assignee、maintainer 或 `trusted_responders`，且原 ticket 仍带完整的失败标。 |
| 每个 tick 都立刻返回 `budget_exhausted` | 当日预算真的用完了(`loopctl budget report` 看是谁),或某个模型的上限配小了。它不是故障:P0 在唤起任何角色之前就停住,一个 token 都没多花。 |
| `by_role[]` 角色不符合预期               | 检查 runner 传给 `callRole` 的 role；SDK usage 不再从历史日志反推。 |

### 还是不行

`.loop/runtime/STATE.md` 是给人读的现状视图;`.loop/runtime/tasks/<ticket>/reports/` 是
那张票的全部证据。带着这两样对照 `.loop/policy/` 与对应角色的 prompt——每条判据在那里
都写着它为什么存在。
