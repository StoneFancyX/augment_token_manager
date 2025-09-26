from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.core.config import settings
from app.core.database import init_db, close_db, get_db
from app.api.auth import router as auth_router
from app.api.tokens import router as tokens_router
from app.api.ide import router as ide_router
from app.services.user_service import UserService
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(_: FastAPI):
    """应用生命周期管理"""
    # 启动时执行
    logger.info("Starting Augment Token Manager...")

    # 初始化数据库
    init_db()

    # 创建默认管理员用户
    db = next(get_db())
    try:
        user_service = UserService(db)
        user_service.create_admin_user(
            settings.ADMIN_USERNAME,
            settings.ADMIN_PASSWORD
        )
        logger.info(f"Admin user '{settings.ADMIN_USERNAME}' created/verified")
    except Exception as e:
        logger.error(f"Failed to create admin user: {e}")
    finally:
        db.close()

    logger.info("Application startup complete")

    yield

    # 关闭时执行
    logger.info("Shutting down...")
    close_db()
    logger.info("Application shutdown complete")


# 创建FastAPI应用
app = FastAPI(
    title=settings.APP_NAME,
    version=settings.VERSION,
    description="Augment Token Manager API",
    lifespan=lifespan,
    redirect_slashes=False,  # 禁用自动重定向
)

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册路由
app.include_router(auth_router, prefix="/api/auth", tags=["认证"])
app.include_router(tokens_router, prefix="/api/tokens", tags=["Token管理"])
app.include_router(ide_router, prefix="/api/ide", tags=["IDE集成"])


# 旧的事件处理器已移除，使用lifespan替代


# 健康检查端点
@app.get("/health")
def health_check():
    """健康检查"""
    return {
        "status": "ok",
        "message": "Augment Token Manager is running",
        "version": settings.VERSION,
    }


# 根路径
@app.get("/")
def root():
    """根路径"""
    return {
        "message": "Welcome to Augment Token Manager API",
        "version": settings.VERSION,
        "docs": "/docs",
    }
