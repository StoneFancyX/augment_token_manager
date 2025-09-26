# 前端依赖安装指南

使用以下命令安装前端依赖：

```bash
cd frontend

# 使用Vite创建Vue + TypeScript项目
pnpm create vite . --template vue-ts

# 安装依赖
pnpm install

# 添加路由和状态管理
pnpm add vue-router@4
pnpm add pinia

# 添加UI组件库
pnpm add ant-design-vue
pnpm add @ant-design/icons-vue

# 添加工具库
pnpm add axios
pnpm add dayjs
# pnpm add vue-sonner  # 可选的消息提示组件

# 添加开发依赖
pnpm add -D @types/node

# Ant Design Vue 不需要额外配置
```

安装完成后，package.json和相关配置文件将自动生成。
