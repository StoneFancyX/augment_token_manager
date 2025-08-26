# Augment Token Manager

一个基于 Go 语言的跨平台 Augment Token 管理系统，提供简洁的 Web 界面来管理您的 Augment Token。

## 功能特性

- ✅ **数据库连接**: 连接 PostgreSQL 数据库并读取现有 Token 数据
- ✅ **列表式界面**: 高效的表格布局显示 Token 信息
- ✅ **智能解析**: 自动解析 portal_info JSON 字段，显示过期时间和剩余次数
- ✅ **模态框操作**: 添加和编辑 Token 的弹窗表单
- ✅ **通知系统**: 实时操作反馈和状态提示
- ✅ **复制功能**: 一键复制 Token 到剪贴板
- 🔄 **Token 管理**: 添加、删除、编辑 Augment Token（部分实现）
- 🌐 **响应式设计**: 适配桌面和移动设备的界面
- 🔒 **数据持久化**: 所有数据存储在 PostgreSQL 数据库中
- 🖥️ **跨平台**: 支持 Windows、Linux 和 macOS

## 技术栈

- **后端**: Go 语言 + Gin Web 框架
- **前端**: HTML5 + CSS3（纯前端实现）
- **数据库**: PostgreSQL
- **数据库驱动**: lib/pq


## 数据库配置

系统连接到以下 PostgreSQL 数据库：

- **数据库名称**: postgres
- **用户名**: postgres
- **密码**: postgres
- **IP 地址**: localhost
- **端口**: 5432

## 安装和运行

### 前置要求

- Go 1.21 或更高版本
- PostgreSQL 数据库（已配置并运行）

### 安装步骤

1. 克隆项目到本地：
```bash
git clone <repository-url>
cd augment_token_manager
```

2. 下载依赖：
```bash
go mod tidy
```

3. 运行应用程序：
```bash
go run cmd/main.go
```

4. 打开浏览器访问：
```
http://localhost:8080
```

## 数据库表结构

系统会自动创建以下数据表：

### tokens 表
```sql
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    token_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API 端点

- `GET /` - 主页面，显示 Token 列表（列表式布局）
- `GET /api/tokens` - 获取所有 Token 的 JSON 数据
- `GET /api/tokens/:id` - 获取单个 Token 的详细信息
- `GET /health` - 健康检查端点

## 开发状态

### ✅ 已完成
- [x] 项目结构搭建
- [x] 数据库连接和表初始化
- [x] Token 数据模型定义
- [x] 数据访问层实现
- [x] 列表式 Web 界面
- [x] Token 列表显示功能
- [x] 模态框添加/编辑表单
- [x] 通知系统
- [x] 复制 Token 功能
- [x] 响应式设计
- [x] portal_info JSON 解析（过期时间、剩余次数）
- [x] 单个 Token 详情 API

### 🔄 待实现
- [ ] 添加新 Token 后端功能
- [ ] 删除 Token 后端功能
- [ ] 编辑 Token 后端功能
- [ ] 一键使用 Token 功能
- [ ] 搜索和过滤功能
- [ ] 批量操作功能

## 版权信息

© 2025 KleinerSource. All rights reserved.

## 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。
