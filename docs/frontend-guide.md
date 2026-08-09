# 前端开发指南（KT 标准）

> 决策：`adr/0002-frontend-stack.md`（shadcn 线）、`adr/0003-admin-stack-refine-antd.md`（后台线）；
> 深度规范：`frontend-stack` skill（v2，按产品类型路由）。
> 新前端项目从 `KTAIorg/kt-frontend` 的 starter 起步，不从零、不重新选型。

## 先按产品类型路由（选型主轴）

| 产品类型 | 栈 | starter |
|---|---|---|
| **后台管理系统**（RBAC/CRUD/表单重，要开箱即用） | **Refine + Ant Design** | `kt-frontend/starters/admin-refine-antd` |
| **新产品 / 纯 AI 新设计**（如 KT Workbench） | shadcn/ui + Tailwind（Vite/Next） | `kt-frontend/starters/web-shadcn` |
| **官网 / 营销**（SEO） | Next + shadcn/ui + Tailwind | 同上 |
| **移动 / TG WebApp** | Vite + shadcn | `kt-frontend/starters/webapp-mobile` |

**口诀**：确定是后台管理系统 → **Refine+antd（吃现成）**；新产品/定制界面 → **shadcn+Tailwind**。

## 共享层（尽量一致 = AI 友好）

React + TypeScript + pnpm；认证一律对接 kt-identity。
- 后台线：Refine（authProvider/accessControlProvider/dataProvider）+ antd（ProTable/ProForm）。
- shadcn 线：Tailwind + shadcn/ui + lucide-react。

> 界面美观取决于设计层（antd 主题 / shadcn 设计 token），与框架无关。

## 起步

### 纯客户端（默认，Vite）
```bash
pnpm create vite@latest <name> -- --template react-ts
cd <name> && pnpm add -D tailwindcss @tailwindcss/vite && pnpm add react-router-dom lucide-react
pnpm dlx shadcn@latest init
pnpm dlx shadcn@latest add button card input
```

### 需 SSR/SEO（官网类，Next）
```bash
pnpm create next-app@latest <name> --typescript --tailwind --app --src-dir --use-pnpm
cd <name> && pnpm dlx shadcn@latest init && pnpm dlx shadcn@latest add button card input
```

## 目录结构

```
src/
├── app/            # 页面 + api/ 路由（App Router）
│   ├── api/
│   ├── layout.tsx
│   └── globals.css # Tailwind 入口
├── components/
│   ├── ui/         # shadcn 组件（进仓库）
│   └── <feature>/
├── lib/            # auth.ts(接 identity) / api client / 工具
└── hooks/
```

## 认证（统一）

- 登录对接 **kt-identity**，不自造。
- 面向客户前端优先 **Telegram 一键登录**（identity client-login challenges + 官方 bot 确认）。
- 内部后台可用 casdoor 账密/扫码。
- 会话逻辑集中在 `src/lib/auth.ts`。

## 规范要点

- 组件用函数式 + hooks；避免 class 组件。
- 样式用 Tailwind utility，避免另写 CSS 文件（globals.css 只放 Tailwind 指令和少量全局）。
- 类型优先：props/API 响应都定 TS 类型。
- Server Components 默认，需交互再标 `"use client"`。
- 数据获取优先 Server Component / Route Handler，避免客户端裸 fetch 泄露逻辑。

## 例外（不适用标准栈）
- Vue 存量（kt-admin-web、kt-desktop）：维护不重写。
- 上游 fork（kt-im-web）：跟随上游。
