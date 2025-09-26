# Augment Token Manager (Python版本)

一个基于 Python + Vue.js 的前后端分离 Augment Token 管理系统，支持 Docker 部署。

## 技术栈

### 后端
- **框架**: FastAPI
- **数据库**: MySQL
- **ORM**: SQLAlchemy
- **认证**: JWT
- **包管理**: uv

### 前端
- **框架**: Vue 3
- **语言**: TypeScript
- **样式**: TailwindCSS
- **构建工具**: Vite
- **包管理**: pnpm

### 部署
- **容器化**: Docker + Docker Compose
- **数据库**: MySQL 8.0

## 功能特性

- ✅ **前后端分离**: 独立的API服务和前端应用
- ✅ **JWT认证**: 安全的用户认证系统
- ✅ **Token管理**: 完整的CRUD操作
- ✅ **状态验证**: 自动验证Token有效性
- ✅ **批量操作**: 支持批量导入和删除
- ✅ **响应式设计**: 适配各种设备
- ✅ **Docker部署**: 一键部署到任何环境


## 项目结构

```
.
├── backend/          # Python后端
│   ├── app/         # 应用代码
│   ├── requirements.txt
│   └── Dockerfile
├── frontend/        # Vue前端
│   ├── src/        # 源代码
│   ├── package.json
│   └── Dockerfile
├── docker/         # Docker配置
├── scripts/        # 部署脚本
└── docker-compose.yml
```

## 快速开始

### 自动化安装 (推荐)

```bash
# 克隆项目
git clone <repository-url>
cd augment_token_manager

# 运行安装脚本
chmod +x scripts/setup.sh
./scripts/setup.sh

# 启动开发环境
chmod +x scripts/deploy.sh
./scripts/deploy.sh dev
```

### 使用 Docker Compose

```bash
# 开发环境
./scripts/deploy.sh dev

# 生产环境
./scripts/deploy.sh prod

# 停止服务
./scripts/deploy.sh stop

# 查看日志
./scripts/deploy.sh logs

# 清理
./scripts/deploy.sh cleanup
```

### 本地开发

#### 前置要求

- Python 3.13+
- Node.js 18+
- uv (Python包管理器)
- pnpm (Node.js包管理器)
- Docker & Docker Compose

#### 后端开发

```bash
cd backend

# 参考 uv-setup.md 安装依赖
# 启动开发服务器
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

#### 前端开发

```bash
cd frontend

# 参考 pnpm-setup.md 安装依赖
# 启动开发服务器
pnpm dev
```

## 环境变量

### 后端环境变量

```bash
# 数据库配置
DATABASE_URL=mysql+aiomysql://user:password@localhost:3306/augment_tokens

# JWT配置
SECRET_KEY=your-secret-key
ACCESS_TOKEN_EXPIRE_MINUTES=30

# 服务器配置
HOST=0.0.0.0
PORT=8000
DEBUG=false
```

### 前端环境变量

```bash
# API地址
VITE_API_BASE_URL=http://localhost:8000
```

## API文档

启动后端服务后，访问以下地址查看API文档：

- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## 数据库表结构

### tokens 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(36) | 主键，UUID |
| tenant_url | TEXT | 租户URL |
| access_token | TEXT | 访问令牌 |
| portal_url | TEXT | 门户URL |
| email_note | TEXT | 邮箱备注 |
| ban_status | VARCHAR(50) | 封禁状态 |
| portal_info | JSON | 门户信息 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

## 开发指南

### 后端开发

1. 使用 FastAPI 框架
2. 遵循 RESTful API 设计原则
3. 使用 SQLAlchemy 进行数据库操作
4. 使用 Pydantic 进行数据验证
5. 使用 JWT 进行身份认证

### 前端开发

1. 使用 Vue 3 Composition API
2. 使用 TypeScript 进行类型检查
3. 使用 TailwindCSS 进行样式设计
4. 使用 Pinia 进行状态管理
5. 使用 Vue Router 进行路由管理

## 版权信息

© 2025 KleinerSource. All rights reserved.

## 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。
