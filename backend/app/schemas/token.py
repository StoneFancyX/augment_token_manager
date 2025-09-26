from pydantic import BaseModel, HttpUrl
from typing import Optional, Any, Dict
from enum import Enum


class TokenStatus(str, Enum):
    """Token状态枚举 - 按照0.6.0版本"""
    ACTIVE = "ACTIVE"
    SUSPENDED = "SUSPENDED"
    INVALID_TOKEN = "INVALID_TOKEN"
    UNAUTHORIZED = "UNAUTHORIZED"
    FORBIDDEN = "FORBIDDEN"
    RATE_LIMITED = "RATE_LIMITED"
    SERVER_ERROR = "SERVER_ERROR"
    UNKNOWN_ERROR = "UNKNOWN_ERROR"
    # 保留旧状态以兼容
    USAGE_LIMIT = "USAGE_LIMIT"
    EXHAUSTED = "EXHAUSTED"
    EXPIRED = "EXPIRED"
    INVALID = "INVALID"


class TokenBase(BaseModel):
    """Token基础模式"""
    tenant_url: str
    access_token: str
    portal_url: str
    email_note: Optional[str] = None


class TokenCreate(TokenBase):
    """创建Token模式"""
    max_usage: Optional[int] = None


class TokenUpdate(BaseModel):
    """更新Token模式"""
    tenant_url: Optional[str] = None
    access_token: Optional[str] = None
    portal_url: Optional[str] = None
    email_note: Optional[str] = None
    max_usage: Optional[int] = None


class TokenResponse(TokenBase):
    """Token响应模式"""
    id: str
    ban_status: Optional[str] = None
    portal_info: Optional[Dict[str, Any]] = None
    usage_count: int = 0
    max_usage: Optional[int] = None
    status_display: str
    is_exhausted: bool
    created_at: str
    updated_at: str

    class Config:
        from_attributes = True


class TokenValidationResult(BaseModel):
    """Token验证结果模式"""
    is_valid: bool
    message: str
    token: TokenResponse


class TokenImportRequest(BaseModel):
    """Token批量导入请求模式"""
    tokens: list[TokenCreate]


class TokenImportResponse(BaseModel):
    """Token批量导入响应模式"""
    success_count: int
    failed_count: int
    tokens: list[TokenResponse]
    errors: list[str]
