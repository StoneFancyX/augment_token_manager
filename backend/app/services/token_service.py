from typing import Optional, List
from sqlalchemy.orm import Session
from sqlalchemy import select, delete
from app.models.token import Token
from app.schemas.token import TokenCreate, TokenUpdate
import uuid


class TokenService:
    """Token服务类"""
    
    def __init__(self, db: Session):
        self.db = db
    
    def get_tokens(self, skip: int = 0, limit: int = 100) -> List[Token]:
        """获取Token列表"""
        result = self.db.execute(
            select(Token)
            .order_by(Token.created_at.desc())
            .offset(skip)
            .limit(limit)
        )
        return result.scalars().all()
    
    def get_all_tokens(self) -> List[Token]:
        """获取所有Token（不分页）"""
        result = self.db.execute(
            select(Token).order_by(Token.created_at.desc())
        )
        return result.scalars().all()
    
    def get_token_by_id(self, token_id: str) -> Optional[Token]:
        """根据ID获取Token"""
        result = self.db.execute(select(Token).where(Token.id == token_id))
        return result.scalar_one_or_none()
    
    def create_token(self, token_data: TokenCreate) -> Token:
        """创建Token"""
        token = Token(
            id=str(uuid.uuid4()),
            tenant_url=token_data.tenant_url,
            access_token=token_data.access_token,
            portal_url=token_data.portal_url,
            email_note=token_data.email_note,
        )
        
        self.db.add(token)
        self.db.commit()
        self.db.refresh(token)
        return token
    
    def update_token(self, token_id: str, token_data: TokenUpdate) -> Optional[Token]:
        """更新Token"""
        token = self.get_token_by_id(token_id)
        if not token:
            return None
        
        # 更新字段
        if token_data.tenant_url is not None:
            token.tenant_url = token_data.tenant_url
        if token_data.access_token is not None:
            token.access_token = token_data.access_token
        if token_data.portal_url is not None:
            token.portal_url = token_data.portal_url
        if token_data.email_note is not None:
            token.email_note = token_data.email_note
        
        self.db.commit()
        self.db.refresh(token)
        return token
    
    def delete_token(self, token_id: str) -> bool:
        """删除Token"""
        result = self.db.execute(delete(Token).where(Token.id == token_id))
        self.db.commit()
        return result.rowcount > 0
    
    def update_token_status(self, token_id: str, ban_status: Optional[str]) -> Optional[Token]:
        """更新Token状态"""
        token = self.get_token_by_id(token_id)
        if not token:
            return None
        
        token.ban_status = ban_status
        self.db.commit()
        self.db.refresh(token)
        return token
    
    def update_token_portal_info(self, token_id: str, portal_info: dict) -> Optional[Token]:
        """更新Token门户信息"""
        token = self.get_token_by_id(token_id)
        if not token:
            return None

        token.portal_info = portal_info
        self.db.commit()
        self.db.refresh(token)
        return token

    def increment_usage_count(self, token_id: str) -> Optional[Token]:
        """增加Token使用次数"""
        token = self.get_token_by_id(token_id)
        if not token:
            return None

        token.usage_count += 1
        self.db.commit()
        self.db.refresh(token)
        return token

    def set_max_usage(self, token_id: str, max_usage: int) -> Optional[Token]:
        """设置Token最大使用次数"""
        token = self.get_token_by_id(token_id)
        if not token:
            return None

        token.max_usage = max_usage
        self.db.commit()
        self.db.refresh(token)
        return token
    
    def create_tokens_batch(self, tokens_data: List[TokenCreate]) -> List[Token]:
        """批量创建Token"""
        tokens = []
        for token_data in tokens_data:
            token = Token(
                id=str(uuid.uuid4()),
                tenant_url=token_data.tenant_url,
                access_token=token_data.access_token,
                portal_url=token_data.portal_url,
                email_note=token_data.email_note,
            )
            tokens.append(token)
            self.db.add(token)
        
        self.db.commit()
        
        # 刷新所有token
        for token in tokens:
            self.db.refresh(token)
        
        return tokens
