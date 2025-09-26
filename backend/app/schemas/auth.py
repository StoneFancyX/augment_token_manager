from pydantic import BaseModel, EmailStr
from typing import Optional


class UserBase(BaseModel):
    """用户基础模式"""
    username: str
    email: Optional[EmailStr] = None


class UserCreate(UserBase):
    """创建用户模式"""
    password: str


class UserUpdate(BaseModel):
    """更新用户模式"""
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    password: Optional[str] = None
    is_active: Optional[bool] = None


class UserResponse(UserBase):
    """用户响应模式"""
    id: str
    is_active: bool
    is_superuser: bool
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class Token(BaseModel):
    """令牌模式"""
    access_token: str
    token_type: str


class TokenData(BaseModel):
    """令牌数据模式"""
    username: Optional[str] = None


class LoginRequest(BaseModel):
    """登录请求模式"""
    username: str
    password: str


class LoginResponse(BaseModel):
    """登录响应模式"""
    access_token: str
    token_type: str
    user: UserResponse
