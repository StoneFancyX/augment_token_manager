from sqlalchemy import Column, String, DateTime, func
from sqlalchemy.ext.declarative import declared_attr
from app.core.database import Base
import uuid


class TimestampMixin:
    """时间戳混入类"""
    
    @declared_attr
    def created_at(cls):
        return Column(DateTime, default=func.now(), nullable=False)
    
    @declared_attr
    def updated_at(cls):
        return Column(DateTime, default=func.now(), onupdate=func.now(), nullable=False)


class UUIDMixin:
    """UUID主键混入类"""
    
    @declared_attr
    def id(cls):
        return Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
