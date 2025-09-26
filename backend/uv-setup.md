# 后端依赖安装指南

使用以下命令安装后端依赖：

```bash
cd backend

# 初始化uv项目
uv init --python 3.13

# 添加核心依赖
uv add fastapi "uvicorn[standard]" sqlalchemy pymysql alembic pydantic pydantic-settings "python-jose[cryptography]" "passlib[bcrypt]" python-multipart httpx python-dotenv

# 添加开发依赖
uv add --dev pytest
uv add --dev pytest-asyncio
uv add --dev pytest-cov
uv add --dev black
uv add --dev isort
uv add --dev flake8
uv add --dev mypy
uv add --dev pre-commit

# 初始化Alembic
uv run alembic init alembic
```

安装完成后，pyproject.toml和alembic配置将自动生成。
