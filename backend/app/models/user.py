from sqlalchemy import Column, String, Boolean
from app.core.database import Base
from app.models.base import UUIDMixin, TimestampMixin


class User(Base, UUIDMixin, TimestampMixin):
    """用户数据模型"""
    
    __tablename__ = "users"
    
    # 用户名
    username = Column(String(100), unique=True, nullable=False, comment="用户名")
    
    # 邮箱
    email = Column(String(255), unique=True, nullable=True, comment="邮箱")
    
    # 密码哈希
    password_hash = Column(String(255), nullable=False, comment="密码哈希")
    
    # 是否激活
    is_active = Column(Boolean, default=True, nullable=False, comment="是否激活")
    
    # 是否为超级用户
    is_superuser = Column(Boolean, default=False, nullable=False, comment="是否为超级用户")
    
    def __repr__(self) -> str:
        return f"<User(id={self.id}, username={self.username})>"
    
    def to_dict(self) -> dict:
        """转换为字典"""
        return {
            "id": self.id,
            "username": self.username,
            "email": self.email,
            "is_active": self.is_active,
            "is_superuser": self.is_superuser,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
