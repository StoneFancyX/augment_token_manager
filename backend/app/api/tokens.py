from typing import List, Annotated
from fastapi import APIRouter, Depends, HTTPException, status, Query, UploadFile, File
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
import json
from app.core.deps import get_current_active_user, get_db
from app.models.user import User
from app.services.token_service import TokenService
from app.services.token_validation_service import TokenValidationService
from app.schemas.token import (
    TokenCreate,
    TokenUpdate,
    TokenResponse,
    TokenValidationResult,
    TokenImportRequest,
    TokenImportResponse,
)

router = APIRouter()


@router.get("", response_model=List[TokenResponse])
def get_tokens(
    skip: int = Query(0, ge=0, description="跳过的记录数"),
    limit: int = Query(100, ge=1, le=1000, description="返回的记录数"),
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """获取Token列表"""
    token_service = TokenService(db)
    tokens = token_service.get_tokens(skip=skip, limit=limit)
    return [TokenResponse(**token.to_dict()) for token in tokens]


@router.get("/statistics")
def get_token_stats(
    current_user: Annotated[User, Depends(get_current_active_user)],
    db: Session = Depends(get_db)
):
    """获取Token统计信息"""
    token_service = TokenService(db)
    tokens = token_service.get_all_tokens()

    if not tokens:
        return {
            "total_tokens": 0,
            "total_credits": 0,
            "available_credits": 0,
            "unlimited_tokens": 0,
            "expired_tokens": 0,
            "valid_tokens": 0
        }

    total_tokens = len(tokens)
    total_credits = 0
    available_credits = 0
    unlimited_tokens = 0
    expired_tokens = 0
    valid_tokens = 0

    # 定义封禁状态列表（按照0.6.0版本的is_banned=true状态）
    banned_statuses = [
        'SUSPENDED', 'UNAUTHORIZED', 'FORBIDDEN', 'UNKNOWN_ERROR', 'INVALID',
        'INVALID_TOKEN', 'EXPIRED'  # 这些状态的token不应计入总额度
    ]

    for token in tokens:
        # 统计有效Token（非封禁状态）
        if not token.ban_status or token.ban_status not in banned_statuses:
            valid_tokens += 1

        # 统计过期Token
        if token.ban_status == 'EXPIRED':
            expired_tokens += 1

        # 统计额度信息（只统计非封禁状态的token）
        if token.portal_info and (not token.ban_status or token.ban_status not in banned_statuses):
            try:
                # 处理portal_info，可能是dict或JSON字符串
                portal_data = token.portal_info
                if isinstance(portal_data, str):
                    import json
                    portal_data = json.loads(portal_data)

                if isinstance(portal_data, dict):
                    credits_balance = portal_data.get('credits_balance', 0)
                    subscription_info = portal_data.get('subscription_info', {})

                    # 检查是否为无限制计划
                    if subscription_info and subscription_info.get('plan_type') == 'unlimited':
                        unlimited_tokens += 1
                        # 无限制计划不计入总额度
                    else:
                        # 累加总额度和可用额度
                        if credits_balance and credits_balance > 0:
                            total_credits += credits_balance
                            available_credits += credits_balance
            except (json.JSONDecodeError, TypeError, AttributeError) as e:
                # 忽略解析错误的数据
                continue

    # 添加调试信息
    debug_info = {
        "processed_tokens": len(tokens),
        "tokens_with_portal_info": sum(1 for token in tokens if token.portal_info),
        "sample_portal_info": tokens[0].portal_info if tokens else None
    }

    return {
        "total_tokens": total_tokens,
        "total_credits": total_credits,
        "available_credits": available_credits,
        "unlimited_tokens": unlimited_tokens,
        "expired_tokens": expired_tokens,
        "valid_tokens": valid_tokens,
        "debug": debug_info
    }


@router.get("/{token_id}", response_model=TokenResponse)
def get_token(
    token_id: str,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """获取单个Token"""
    token_service = TokenService(db)
    token = token_service.get_token_by_id(token_id)
    
    if not token:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Token不存在"
        )
    
    return TokenResponse(**token.to_dict())


@router.post("", response_model=TokenResponse)
def create_token(
    token_data: TokenCreate,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """创建Token"""
    token_service = TokenService(db)
    
    try:
        token = token_service.create_token(token_data)
        return TokenResponse(**token.to_dict())
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"创建Token失败: {str(e)}"
        )


@router.put("/{token_id}", response_model=TokenResponse)
def update_token(
    token_id: str,
    token_data: TokenUpdate,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """更新Token"""
    token_service = TokenService(db)
    
    try:
        token = token_service.update_token(token_id, token_data)
        if not token:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Token不存在"
            )
        return TokenResponse(**token.to_dict())
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"更新Token失败: {str(e)}"
        )


@router.delete("/{token_id}")
def delete_token(
    token_id: str,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """删除Token"""
    token_service = TokenService(db)
    
    success = token_service.delete_token(token_id)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Token不存在"
        )
    
    return {"message": "Token删除成功"}


@router.post("/{token_id}/validate", response_model=TokenValidationResult)
def validate_token(
    token_id: str,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """验证Token状态"""
    token_service = TokenService(db)
    validation_service = TokenValidationService(db)
    
    token = token_service.get_token_by_id(token_id)
    if not token:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Token不存在"
        )
    
    try:
        result = validation_service.validate_token(token)
        return result
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"验证Token失败: {str(e)}"
        )


@router.post("/{token_id}/refresh", response_model=TokenResponse)
def refresh_token(
    token_id: str,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """刷新Token信息"""
    token_service = TokenService(db)
    validation_service = TokenValidationService(db)
    
    token = token_service.get_token_by_id(token_id)
    if not token:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Token不存在"
        )
    
    try:
        updated_token = validation_service.refresh_token_info(token)
        return TokenResponse(**updated_token.to_dict())
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"刷新Token失败: {str(e)}"
        )


@router.post("/import", response_model=TokenImportResponse)
def import_tokens(
    import_data: TokenImportRequest,
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """批量导入Token"""
    token_service = TokenService(db)
    
    created_tokens = []
    errors = []
    
    for token_data in import_data.tokens:
        try:
            token = token_service.create_token(token_data)
            created_tokens.append(TokenResponse(**token.to_dict()))
        except Exception as e:
            errors.append(f"创建Token失败: {str(e)}")
    
    return TokenImportResponse(
        success_count=len(created_tokens),
        failed_count=len(errors),
        tokens=created_tokens,
        errors=errors
    )


@router.post("/batch-delete")
def batch_delete_tokens(
    token_ids: List[str],
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """批量删除Token"""
    token_service = TokenService(db)
    
    success_count = 0
    failed_count = 0
    errors = []
    
    for token_id in token_ids:
        try:
            success = token_service.delete_token(token_id)
            if success:
                success_count += 1
            else:
                failed_count += 1
                errors.append(f"Token {token_id} 不存在")
        except Exception as e:
            failed_count += 1
            errors.append(f"删除Token {token_id} 失败: {str(e)}")
    
    return {
        "success_count": success_count,
        "failed_count": failed_count,
        "errors": errors,
        "message": f"批量删除完成，成功: {success_count}，失败: {failed_count}"
    }


@router.post("/batch-validate")
def batch_validate_tokens(
    token_ids: List[str],
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """批量验证Token"""
    token_service = TokenService(db)
    validation_service = TokenValidationService(db)

    results = []
    success_count = 0
    failed_count = 0

    for token_id in token_ids:
        try:
            token = token_service.get_token_by_id(token_id)
            if not token:
                results.append({
                    "token_id": token_id,
                    "is_valid": False,
                    "message": "Token不存在",
                    "error": True
                })
                failed_count += 1
                continue

            result = validation_service.validate_token(token)
            results.append({
                "token_id": token_id,
                "is_valid": result.is_valid,
                "message": result.message,
                "error": False
            })
            
            if result.is_valid:
                success_count += 1
            else:
                failed_count += 1
                
        except Exception as e:
            results.append({
                "token_id": token_id,
                "is_valid": False,
                "message": str(e),
                "error": True
            })
            failed_count += 1

    return {
        "results": results,
        "success_count": success_count,
        "failed_count": failed_count,
        "message": f"批量验证完成，有效: {success_count}，无效: {failed_count}"
    }


@router.post("/batch-refresh")
def batch_refresh_tokens(
    current_user: Annotated[User, Depends(get_current_active_user)] = None,
    db: Annotated[Session, Depends(get_db)] = None,
):
    """批量刷新所有Token信息"""
    token_service = TokenService(db)
    validation_service = TokenValidationService(db)
    
    # 获取所有Token
    tokens = token_service.get_all_tokens()
    
    if not tokens:
        return {
            "message": "没有找到任何Token",
            "success_count": 0,
            "failed_count": 0,
            "results": []
        }
    
    results = []
    success_count = 0
    failed_count = 0
    
    for token in tokens:
        try:
            # 验证并刷新Token状态
            result = validation_service.validate_token(token)
            results.append({
                "token_id": token.id,
                "email_note": token.email_note,
                "status": "success",
                "is_valid": result.is_valid,
                "message": result.message
            })
            success_count += 1
        except Exception as e:
            results.append({
                "token_id": token.id,
                "email_note": token.email_note,
                "status": "failed",
                "error": str(e)
            })
            failed_count += 1
    
    return {
        "message": f"批量刷新完成，成功 {success_count} 个，失败 {failed_count} 个",
        "success_count": success_count,
        "failed_count": failed_count,
        "results": results
    }


@router.post("/export")
async def export_tokens(
    token_ids: List[int],
    current_user: Annotated[User, Depends(get_current_active_user)],
    db: Session = Depends(get_db)
):
    """导出Token数据为JSON"""
    try:
        token_service = TokenService(db)

        # 获取指定的tokens
        tokens = []
        for token_id in token_ids:
            token = token_service.get_token(token_id)
            if token:
                tokens.append({
                    "tenant_url": token.tenant_url,
                    "access_token": token.access_token,
                    "portal_url": token.portal_url,
                    "email_note": token.email_note
                })

        if not tokens:
            raise HTTPException(status_code=404, detail="未找到要导出的Token")

        # 返回JSON响应
        return JSONResponse(
            content=tokens,
            headers={
                "Content-Disposition": "attachment; filename=tokens_export.json"
            }
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"导出失败: {str(e)}")


@router.post("/import")
async def import_tokens(
    current_user: Annotated[User, Depends(get_current_active_user)],
    db: Session = Depends(get_db),
    file: UploadFile = File(...)
):
    """从JSON文件导入Token数据"""
    try:
        # 检查文件类型
        if not file.filename.endswith('.json'):
            raise HTTPException(status_code=400, detail="只支持JSON文件")

        # 读取文件内容
        content = await file.read()
        data = json.loads(content.decode('utf-8'))

        if not isinstance(data, list):
            raise HTTPException(status_code=400, detail="JSON文件格式错误，应为数组格式")

        token_service = TokenService(db)
        success_count = 0
        failed_count = 0
        results = []

        for item in data:
            try:
                # 验证必需字段
                required_fields = ["tenant_url", "access_token", "portal_url", "email_note"]
                for field in required_fields:
                    if field not in item:
                        raise ValueError(f"缺少必需字段: {field}")

                # 创建Token
                token_create = TokenCreate(
                    tenant_url=item["tenant_url"],
                    access_token=item["access_token"],
                    portal_url=item["portal_url"],
                    email_note=item["email_note"]
                )

                token = token_service.create_token(token_create)
                results.append({
                    "success": True,
                    "token_id": token.id,
                    "email_note": token.email_note
                })
                success_count += 1

            except Exception as e:
                results.append({
                    "success": False,
                    "error": str(e),
                    "data": item
                })
                failed_count += 1

        return {
            "message": f"导入完成，成功 {success_count} 个，失败 {failed_count} 个",
            "success_count": success_count,
            "failed_count": failed_count,
            "results": results
        }

    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="JSON文件格式错误")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"导入失败: {str(e)}")
