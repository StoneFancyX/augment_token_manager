from sqlalchemy import Column, String, Text, JSON, Integer, Enum
from app.core.database import Base
from app.models.base import UUIDMixin, TimestampMixin
import enum


class TokenStatus(str, enum.Enum):
    """Token状态枚举 - 按照0.6.0版本"""
    ACTIVE = "ACTIVE"                    # 正常状态
    SUSPENDED = "SUSPENDED"              # 被暂停/封禁
    INVALID_TOKEN = "INVALID_TOKEN"      # Token无效
    UNAUTHORIZED = "UNAUTHORIZED"        # 未授权
    FORBIDDEN = "FORBIDDEN"              # 禁止访问
    RATE_LIMITED = "RATE_LIMITED"        # 限流
    SERVER_ERROR = "SERVER_ERROR"        # 服务器错误
    UNKNOWN_ERROR = "UNKNOWN_ERROR"      # 未知错误
    # 保留旧状态以兼容
    USAGE_LIMIT = "USAGE_LIMIT"          # 使用次数耗尽
    EXHAUSTED = "EXHAUSTED"              # 已耗尽（同USAGE_LIMIT）
    EXPIRED = "EXPIRED"                  # 已过期
    INVALID = "INVALID"                  # 无效（通用）


class Token(Base, UUIDMixin, TimestampMixin):
    """Token数据模型"""

    __tablename__ = "tokens"

    # 租户URL
    tenant_url = Column(Text, nullable=False, comment="租户URL")

    # 访问令牌
    access_token = Column(Text, nullable=False, comment="访问令牌")

    # 门户URL
    portal_url = Column(Text, nullable=False, comment="门户URL")

    # 邮箱备注
    email_note = Column(Text, nullable=True, comment="邮箱备注")

    # 封禁状态/Token状态
    ban_status = Column(String(50), nullable=True, comment="Token状态")

    # 门户信息 (JSON格式)
    portal_info = Column(JSON, nullable=True, comment="门户信息")

    # 使用次数统计
    usage_count = Column(Integer, default=0, nullable=False, comment="使用次数")

    # 最大使用次数限制
    max_usage = Column(Integer, nullable=True, comment="最大使用次数限制")
    
    def __repr__(self) -> str:
        return f"<Token(id={self.id}, email_note={self.email_note})>"
    
    def to_dict(self) -> dict:
        """转换为字典"""
        return {
            "id": self.id,
            "tenant_url": self.tenant_url,
            "access_token": self.access_token,
            "portal_url": self.portal_url,
            "email_note": self.email_note,
            "ban_status": self.ban_status,
            "portal_info": self.portal_info,
            "usage_count": self.usage_count or 0,
            "max_usage": self.max_usage,
            "status_display": self.status_display,
            "is_exhausted": self.is_exhausted,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }

    @property
    def is_exhausted(self) -> bool:
        """检查是否已耗尽使用次数"""
        if self.max_usage is None:
            return False
        return self.usage_count >= self.max_usage

    @property
    def status_display(self) -> str:
        """获取状态显示文本"""
        # 优先检查ban_status（按照原版逻辑）
        if self.ban_status:
            return self.ban_status

        # 如果没有ban_status，说明验证通过，应该显示为ACTIVE
        # portal_info的余额信息在前端单独显示，不影响状态

        # 只有在使用次数耗尽时才显示EXHAUSTED
        if self.is_exhausted:
            return TokenStatus.EXHAUSTED

        # 默认为正常状态
        return TokenStatus.ACTIVE
