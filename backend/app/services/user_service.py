from typing import Optional
from sqlalchemy.orm import Session
from sqlalchemy import select
from app.models.user import User
from app.schemas.auth import UserCreate, UserUpdate
from app.core.security import get_password_hash, verify_password
import uuid


class UserService:
    """用户服务类"""
    
    def __init__(self, db: Session):
        self.db = db
    
    def get_user_by_id(self, user_id: str) -> Optional[User]:
        """根据ID获取用户"""
        result = self.db.execute(select(User).where(User.id == user_id))
        return result.scalar_one_or_none()
    
    def get_user_by_username(self, username: str) -> Optional[User]:
        """根据用户名获取用户"""
        result = self.db.execute(select(User).where(User.username == username))
        return result.scalar_one_or_none()
    
    def get_user_by_email(self, email: str) -> Optional[User]:
        """根据邮箱获取用户"""
        result = self.db.execute(select(User).where(User.email == email))
        return result.scalar_one_or_none()
    
    def create_user(self, user_data: UserCreate) -> User:
        """创建用户"""
        # 检查用户名是否已存在
        existing_user = self.get_user_by_username(user_data.username)
        if existing_user:
            raise ValueError("用户名已存在")
        
        # 检查邮箱是否已存在（如果提供了邮箱）
        if user_data.email:
            existing_email = self.get_user_by_email(user_data.email)
            if existing_email:
                raise ValueError("邮箱已存在")
        
        # 创建新用户
        user = User(
            id=str(uuid.uuid4()),
            username=user_data.username,
            email=user_data.email,
            password_hash=get_password_hash(user_data.password),
            is_active=True,
            is_superuser=False,
        )
        
        self.db.add(user)
        self.db.commit()
        self.db.refresh(user)
        return user
    
    def update_user(self, user_id: str, user_data: UserUpdate) -> Optional[User]:
        """更新用户"""
        user = self.get_user_by_id(user_id)
        if not user:
            return None
        
        # 更新字段
        if user_data.username is not None:
            # 检查新用户名是否已被其他用户使用
            existing_user = self.get_user_by_username(user_data.username)
            if existing_user and existing_user.id != user_id:
                raise ValueError("用户名已存在")
            user.username = user_data.username
        
        if user_data.email is not None:
            # 检查新邮箱是否已被其他用户使用
            existing_email = self.get_user_by_email(user_data.email)
            if existing_email and existing_email.id != user_id:
                raise ValueError("邮箱已存在")
            user.email = user_data.email
        
        if user_data.password is not None:
            user.password_hash = get_password_hash(user_data.password)
        
        if user_data.is_active is not None:
            user.is_active = user_data.is_active
        
        self.db.commit()
        self.db.refresh(user)
        return user
    
    def delete_user(self, user_id: str) -> bool:
        """删除用户"""
        user = self.get_user_by_id(user_id)
        if not user:
            return False
        
        self.db.delete(user)
        self.db.commit()
        return True
    
    def authenticate_user(self, username: str, password: str) -> Optional[User]:
        """验证用户登录"""
        user = self.get_user_by_username(username)
        if not user:
            return None
        
        if not verify_password(password, user.password_hash):
            return None
        
        return user
    
    def create_admin_user(self, username: str, password: str, email: str = None) -> User:
        """创建管理员用户"""
        # 检查用户名是否已存在
        existing_user = self.get_user_by_username(username)
        if existing_user:
            return existing_user  # 如果管理员已存在，直接返回
        
        # 创建管理员用户
        user = User(
            id=str(uuid.uuid4()),
            username=username,
            email=email,
            password_hash=get_password_hash(password),
            is_active=True,
            is_superuser=True,
        )
        
        self.db.add(user)
        self.db.commit()
        self.db.refresh(user)
        return user
