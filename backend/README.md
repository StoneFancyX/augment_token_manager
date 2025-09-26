# Augment Token Manager Backend

基于 FastAPI 的后端 API 服务。

## 技术栈

- **框架**: FastAPI
- **数据库**: MySQL 8.0
- **ORM**: SQLAlchemy 2.0 (异步)
- **认证**: JWT
- **包管理**: uv
- **数据库迁移**: Alembic

## 项目结构

```
backend/
├── app/
│   ├── api/          # API路由
│   ├── core/         # 核心配置
│   ├── models/       # 数据库模型
│   ├── schemas/      # Pydantic模式
│   ├── services/     # 业务逻辑
│   └── utils/        # 工具函数
├── alembic/          # 数据库迁移
├── pyproject.toml    # 项目配置
└── README.md
```

## 开发环境设置

### 1. 安装依赖

参考 `uv-setup.md` 文件中的详细安装步骤。

```bash
# 初始化项目并安装依赖
uv init --python 3.11
uv add fastapi "uvicorn[standard]" sqlalchemy aiomysql alembic pydantic pydantic-settings "python-jose[cryptography]" "passlib[bcrypt]" python-multipart httpx python-dotenv loguru

# 安装开发依赖
uv add --dev pytest pytest-asyncio pytest-cov black isort flake8 mypy pre-commit
```

### 2. 环境变量配置

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

### 3. 数据库迁移

```bash
# 初始化Alembic
uv run alembic init alembic

# 创建迁移文件
uv run alembic revision --autogenerate -m "Initial migration"

# 执行迁移
uv run alembic upgrade head
```

### 4. 启动开发服务器

```bash
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

## API 文档

启动服务后访问：

- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## 代码质量

### 格式化代码

```bash
uv run black .
uv run isort .
```

### 类型检查

```bash
uv run mypy .
```

### 运行测试

```bash
uv run pytest
```

## Docker 部署

```bash
# 构建镜像
docker build -t augment-backend .

# 运行容器
docker run -p 8000:8000 augment-backend
```
