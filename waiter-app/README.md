# 服务员端 App

基于 uni-app 开发的餐厅服务员端应用，支持 H5、微信小程序和 App 多端运行。

## 功能特性

- 🔐 **账号登录**：使用管理员分配的账号登录系统
- 🪑 **桌位状态**：实时查看桌位状态（空闲/已入座/已下单/已上菜/已结账）
- 🍽️ **点餐下单**：协助顾客点餐、下单、加菜
- 📋 **订单管理**：查看订单列表和订单详情
- ✅ **上菜标记**：菜品制作完成后标记上菜
- 🔄 **退菜功能**：支持订单退菜操作
- 🔔 **呼叫接收**：通过 WebSocket 实时接收顾客呼叫服务员通知
- 📱 **uni-push 推送**：接收新订单提醒（App端）

## 技术栈

- **框架**：uni-app (Vue 3 + TypeScript)
- **状态管理**：Pinia
- **样式**：SCSS
- **实时通信**：WebSocket
- **推送服务**：uni-push

## 项目结构

```
waiter-app/
├── src/
│   ├── pages/              # 页面
│   │   ├── login/          # 登录页
│   │   ├── tables/         # 桌位状态页
│   │   ├── menu/           # 点餐页
│   │   ├── orders/         # 订单列表/详情页
│   │   ├── calls/          # 呼叫记录页
│   │   └── mine/           # 我的页
│   ├── services/           # API 服务
│   │   ├── request.ts      # 请求封装
│   │   ├── auth.ts         # 认证接口
│   │   ├── waiter.ts       # 服务员相关接口
│   │   ├── order.ts        # 订单接口
│   │   ├── product.ts      # 菜品接口
│   │   ├── table.ts        # 桌位接口
│   │   ├── push.ts         # uni-push 推送服务
│   │   └── websocket.ts    # WebSocket 服务
│   ├── store/              # Pinia 状态管理
│   │   ├── user.ts         # 用户状态
│   │   ├── cart.ts         # 购物车状态
│   │   └── websocket.ts    # WebSocket 状态
│   ├── styles/             # 全局样式
│   ├── types/              # TypeScript 类型定义
│   ├── static/             # 静态资源
│   ├── App.vue
│   ├── main.ts
│   ├── pages.json
│   └── manifest.json
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```

## 后端接口依赖

服务员端依赖后端新增的以下接口：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/waiter/stats` | 获取服务员统计数据 |
| GET | `/api/v1/waiter/tables` | 获取桌位状态列表 |
| PUT | `/api/v1/waiter/order-items/cook-status` | 更新菜品制作状态 |
| POST | `/api/v1/waiter/order-items/serve` | 标记菜品已上菜 |
| POST | `/api/v1/waiter/orders/:id/items` | 给订单加菜 |
| GET | `/api/v1/waiter/calls` | 获取呼叫记录 |
| POST | `/api/v1/waiter/calls/:id/handle` | 处理呼叫 |
| POST | `/api/v1/waiter/call` | 顾客呼叫服务员（小程序端调用） |
| GET | `/api/v1/waiter/ws` | WebSocket 连接 |

## 使用说明

### 安装依赖

```bash
cd waiter-app
npm install
```

### 开发运行

```bash
# H5
npm run dev:h5

# 微信小程序
npm run dev:mp-weixin

# App
npm run dev:app
```

### 打包构建

```bash
# H5
npm run build:h5

# 微信小程序
npm run build:mp-weixin

# App
npm run build:app
```

## 配置说明

### API 地址

修改 `src/services/request.ts` 中的 `BASE_URL` 为你的后端服务地址。

### uni-push 配置

在 `src/manifest.json` 中配置 uni-push 的 appid、appkey 和 appsecret。
