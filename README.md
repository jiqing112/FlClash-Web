# FlClash-Web

以 [FlClash](https://github.com/chen08209/FlClash) v0.9.3 为参照,用 **Go + SQLite + Svelte** 复刻的 Web 版客户端。
单个二进制,监听一个 HTTP 端口,内置完整 mihomo(Clash.Meta)核心,在 Windows / Linux / macOS 上行为一致。

## 架构

```
┌────────────────────────────── 单个二进制 flclash-web ──────────────────────────────┐
│                                                                                    │
│  Svelte 前端 (go:embed 内嵌)  ←→  REST API + WebSocket  ←→  hub 业务层             │
│                                        │                        │                  │
│                                        │                 SQLite(flclash.db)        │
│                                        │                 配置 / 配置档案            │
│                                        ▼                                           │
│                              FlClash core(原版 Go core,进程内)                     │
│                          ← 命名管道(Win)/ TCP(Unix)帧协议,与桌面版完全相同 →        │
│                                        │                                           │
│                              mihomo(Clash.Meta fork)                              │
└────────────────────────────────────────────────────────────────────────────────────┘
```

**复刻了原版的三层结构**:

- 前端:页面、卡片、主题、i18n(zh_CN/en)直接取自 FlClash 的 `arb/` 文案;配色用与
  prism_ui 相同的 material-color-utilities 算法(`seededColorScheme` 忠实移植,vivid /
  tinted / soft / monochrome 四种风格 + 纯黑模式);图标按 `app_glyphs.dart` 逐个移植(124 个)。
- 后端 hub:移植了 Dart 侧的 `makeRealProfileTask`(配置覆写管线)、`buildGroups` /
  `computeSort`(策略组计算)、setup/updateConfig 控制流、请求/日志/DNS 缓存、系统代理
  (Windows 注册表 / gsettings / networksetup)。
- 核心:**原封不动使用 FlClash 的 Go core 子模块**(含 Clash.Meta fork),
  走和桌面版一样的 IPC 协议(`config.setup` / `proxy.select` / 延迟测速 / 事件批量推送)。

## 运行

```bash
# 开发:构建前端 + 后端
cd frontend && npm install && npm run build
cd ../backend && go build -o flclash-web.exe .

# 启动(默认 127.0.0.1:8787,数据目录 %AppData%/flclash-web)
./flclash-web.exe -addr 127.0.0.1:8787

# 指定数据目录与监听地址
./flclash-web.exe -addr 0.0.0.0:8787 -home /var/lib/flclash-web
```

浏览器打开 `http://127.0.0.1:8787` 即得完整界面。

Go 需要 1.27+(core 模块要求);`GOTOOLCHAIN=auto` 会自动下载。

## 功能对照

| 功能 | 状态 |
|---|---|
| 仪表盘(状态卡/出站 IP/网速曲线/出站模式/系统代理/TUN/内网 IP/服务状态/配置卡/策略组/连接统计/内存) | ✅ |
| 仪表盘卡片编辑(增删/排序,与原版 widget 列表一致) | ✅ |
| 代理页(tab/list 双模式、节点卡片、延迟测速、选择节点、隐藏超时节点) | ✅ |
| 配置页(URL/文件导入、订阅下载含 userinfo 流量信息、选择/删除/编辑/同步) | ✅ |
| 请求页(实时连接 + 最近请求,速度列,关闭连接) | ✅ |
| DNS 查询、日志页 | ✅ |
| 资源页(Geo 资源更新、外部代理提供者) | ✅ |
| 工具页(含语言/主题/常规/网络设置、关于) | ✅ |
| 主题(明暗/跟随系统、8 种种子色、4 种配色风格、纯黑) | ✅ |
| 中文/英文(文案直接使用原版 arb) | ✅ |
| 侧边栏展开/收起、移动端浮动 dock | ✅ |
| 系统代理(Win 注册表 / macOS networksetup / Linux gsettings) | ✅ |
| 配置覆写(端口/TUN/DNS/hosts/test-url/常规覆写键选择) | ✅ |
| mihomo 全功能(规则分流、TUN、DNS、提供者等,由原版 core 提供) | ✅ |
| 热键、托盘、窗口管理、备份/DAV、MITM、JS 脚本扩展 | ❌ 桌面专属/未实现 |

## 构建(三平台)

```bash
./scripts/build.ps1   # Windows + Linux + macOS amd64/arm64 产物输出到 dist/
./scripts/build.sh    # 同上(bash)
```

产物:`dist/flclash-web-<os>-<arch>[.exe]`,每个都是把前端内嵌进去的单文件。

## 真机对照

前端按 v0.9.3 Windows 实机逐卡重绘:侧栏(220px、灰选中 pill + 3×16 蓝指示条)、
仪表盘 208/64/176 行高网格、cookie-12 blob 装饰、cookie-9 启动按钮、三联出站模式
connected pills(外圆角 24/内 10)、系统代理与虚拟网卡 tile、M3 波浪订阅进度线、
服务状态品牌 logo 探测格、连接表 connected-pill tabs(外圆角 = 高度一半)。
所有参数取自 `lib/views/dashboard/**` 与 `plugins/prism_ui/**` 源码。

**订阅实测**:导入 Kitty Network 订阅 → 34 节点/3 组正确解析 → 订阅信息
(72.4 GB / 512 GB、剩余天数)与真机一致 → 全组延迟测速(马来西亚 70ms)→
经 `curl --proxy` 走 🇲🇾Malaysia 02 出口 204 成功 → 请求/流量曲线实时更新。

## 已知差异

- 旗帜 emoji 在 Windows Chrome 上显示为字母(如 "HK"),原版用 twemoji 图片渲染;
  节点自带的 🇭🇰 字符在支持的平台上正常显示。
- TUN 在 Windows 上需要管理员权限运行本程序(mihomo 自身行为,与原版一致;原版桌面
  通过 helper 服务提权,Web 版暂无该机制)。
- JS 脚本覆写、MITM 模块、备份恢复、热键、系统托盘是桌面专属能力,未实现。

## 目录

```
backend/
  main.go               入口(-addr / -home)
  core/                 FlClash 原版 core(clash.meta fork + FlClash 包装层)
  internal/rpc/         帧协议宿主(与桌面版相同的 JSON RPC)
  internal/hub/         业务层(profile 构建、组计算、系统代理、事件泵)
  internal/store/       SQLite 存储 profiles + config
  internal/api/         REST + WebSocket + 静态资源(dist 内嵌)
frontend/
  src/pages/            Dashboard / Proxies / Profiles / Requests / Dns / Logs / Resources / Tools
  src/lib/              主题(color+MCU)、i18n(arb)、stores、ws、格式化
  src/i18n/             原版 arb 文案(en / zh_CN)
scripts/                跨平台构建脚本
reference/FlClash/      参考仓库(GPL-3.0)
```

## 鸣谢

- [FlClash](https://github.com/chen08209/FlClash) —— 本项目的界面、交互与核心集成方式全部参照 FlClash 复刻,profile 构建管线、策略组计算等业务逻辑由其 Dart 源码移植;感谢作者 [chen08209](https://github.com/chen08209) 与所有贡献者。
- [mihomo(Clash.Meta)](https://github.com/MetaCubeX/mihomo) —— FlClash 内核。
- [material-color-utilities](https://github.com/material-foundation/material-color-utilities) —— 主题配色算法。

## 许可

本项目与 FlClash 一样基于 **[GPL-3.0](LICENSE)** 开源。FlClash 与 mihomo 的代码以各自的开源协议随本仓库分发。
