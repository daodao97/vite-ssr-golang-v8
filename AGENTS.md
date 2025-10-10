# Repository Guidelines

## Project Structure & Module Organization
- 顶层 `main.go` 负责启动嵌入式 SSR 服务器，并通过 `pkg/routematcher` 将请求映射到数据抓取逻辑；`pkg/server.go` 处理 HTTP、SSR 数据注入与 `/__ssr_data` 接口。
- V8 渲染池与 Promise 辅助位于 `pkg/renderer`，如需扩展渲染流程请优先复用 `Renderer.Render` 与 `IsolatePool`。
- 前端代码集中在 `frontend/src`，路由依 `src/pages` 文件自动生成；状态注水与更新可通过 `composables/useSsrData.ts`。
- 构建产物输出到 `dist/client` 与 `dist/server` 并由 `go:embed` 打包，仓库中请保持空目录但勿提交生成文件。

## Build, Test, and Development Commands
- 进入前端目录（默认 `webssr`）执行 `pnpm install`：首次同步依赖；使用 pnpm 锁文件。
- `pnpm dev` 启动 Vite 前端调试（端口 3333）；SSR 行为需通过 `pnpm build` 或 Docker 体验。
- `pnpm build` 顺序执行 `build:client` 与 `build:server`，生成可嵌入到 Go 的 `dist`；若仅需服务器端 bundle，使用 `pnpm build:server`。
- `make build` 或 `go build -o build/server main.go` 将输出精简二进制；构建前确认 dist 已更新。
- `docker compose up -d` 拉起完整示例，`docker compose down` 负责清理容器与网络。

## Coding Style & Naming Conventions
- Go 代码使用 `gofmt`、`goimports`；导出符号 PascalCase，私有符号驼峰式；保持函数短小并于 `pkg/routematcher` 中编写可测试的纯函数。
- 前端遵循项目内 ESLint 规则（`pnpm lint`）与 TypeScript 类型检查（`pnpm typecheck`）；Vue 组件命名使用 PascalCase，路由文件采用参数方括号或短横线。
- 样式通过 UnoCSS 管理，配置位于 `uno.config.ts`，共享样式请归档到 `src/styles`。

## Testing Guidelines
- 前端单测使用 Vitest + jsdom（`pnpm test`）；快照更新需显式运行 `pnpm test -u` 并核对改动。
- 建议为新路由或数据抓取逻辑编写 E2E 或单元测试，可在 `frontend/test` 仓储示例基础上扩展。
- Go 层暂未引入测试，请在 `pkg/...` 下添加 `_test.go` 并使用 `go test ./pkg/...` 验证；对于依赖 V8 的逻辑，优先抽象接口以便 mock。

## Commit & Pull Request Guidelines
- Commit 标题使用祈使句和首字母大写（如 `Add SSR data matcher`），控制在 72 个字符以内；必要时在正文详述背景和验证步骤。
- PR 描述至少包含：目的、关键改动、验证方式（复制命令输出），涉及 UI 或数据行为改动时附上截图或日志片段。
- 提交前确保代码已 `pnpm lint`、`pnpm typecheck`、`pnpm test` 与 `go test ./pkg/...`（如适用），并同步更新相关文档。

## Security & Configuration Tips
- 应用默认监听 `PORT=8080`；部署时可通过环境变量覆盖，Dockerfile 预设 amd64，需要 Apple Silicon 请删去 `--platform`。
- Go 构建依赖 CGO 与 V8，请在 CI 安装 `gcc/g++`；若使用热更新，可结合 `air` 或 `watchexec` 监听重建。
- SSR 数据接口返回 JSON，避免暴露敏感字段；请勿将 `.env`、密钥或临时调试脚本提交到仓库。
