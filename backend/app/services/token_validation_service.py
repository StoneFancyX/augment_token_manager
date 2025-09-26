from typing import Dict, Any
import httpx
from sqlalchemy.orm import Session
from app.models.token import Token
from app.schemas.token import TokenValidationResult, TokenResponse
from app.services.token_service import TokenService
import logging

logger = logging.getLogger(__name__)


class TokenValidationService:
    """Token验证服务类"""
    
    def __init__(self, db: Session):
        self.db = db
        self.token_service = TokenService(db)
    
    def validate_token(self, token: Token) -> TokenValidationResult:
        """验证Token状态"""
        try:
            # 调用外部API验证Token
            # _call_external_api方法内部已经会设置具体的状态
            is_valid = self._call_external_api(token)

            # 获取更新后的Token（状态已在_call_external_api中设置）
            updated_token = self.token_service.get_token_by_id(token.id)

            # 根据验证结果设置消息
            if is_valid:
                message = "Token状态正常"
            else:
                # 根据具体状态设置消息
                if updated_token.ban_status == "SUSPENDED":
                    message = "账号已被封禁"
                elif updated_token.ban_status == "INVALID_TOKEN":
                    message = "Token无效"
                elif updated_token.ban_status in ["UNAUTHORIZED", "FORBIDDEN", "UNKNOWN_ERROR"]:
                    message = "账号已被封禁"
                else:
                    message = "Token状态异常"
            
            return TokenValidationResult(
                is_valid=is_valid,
                message=message,
                token=TokenResponse(**updated_token.to_dict()) if updated_token else TokenResponse(**token.to_dict())
            )
        except Exception as e:
            logger.error(f"验证Token失败: {e}")
            return TokenValidationResult(
                is_valid=False,
                message=f"验证失败: {str(e)}",
                token=TokenResponse(**token.to_dict())
            )
    
    def refresh_token_info(self, token: Token) -> Token:
        """刷新Token信息"""
        try:
            # 获取Token的门户信息
            portal_info = self._get_portal_info(token)
            
            # 更新Token的门户信息
            updated_token = self.token_service.update_token_portal_info(
                token.id, portal_info
            )
            return updated_token
        except Exception as e:
            logger.error(f"刷新Token信息失败: {e}")
            raise
    
    def _call_external_api(self, token: Token) -> bool:
        """调用外部API验证Token - 完全按照0.6.0版本逻辑"""
        try:
            # 构建API URL（按照0.6.0版本逻辑）
            base_url = token.tenant_url.rstrip('/') + '/'
            url = f"{base_url}find-missing"

            # 设置请求头
            headers = {
                "Authorization": f"Bearer {token.access_token}",
                "Content-Type": "application/json"
            }

            # 空请求体（按照0.6.0版本逻辑）
            request_body = {}

            # 发送请求
            with httpx.Client(timeout=30.0) as client:
                response = client.post(url, json=request_body, headers=headers)
                response_text = response.text.lower()
                status_code = response.status_code

                # 按照0.6.0版本的完全一致的逻辑
                # 1. 优先检查响应内容中的"suspended"关键词
                if "suspended" in response_text:
                    self.token_service.update_token_status(token.id, "SUSPENDED")
                    return False
                # 2. 检查"invalid token"关键词
                elif "invalid token" in response_text:
                    self.token_service.update_token_status(token.id, "INVALID_TOKEN")
                    return False
                # 3. 检查成功状态码（200-299）
                elif 200 <= status_code < 300:
                    # 成功状态且无问题关键词 - 账号正常
                    self.token_service.update_token_status(token.id, "ACTIVE")
                    return True
                # 4. 处理其他错误状态码
                else:
                    # 按照0.6.0版本的状态码映射
                    if status_code == 401:
                        self.token_service.update_token_status(token.id, "UNAUTHORIZED")
                        return False
                    elif status_code == 403:
                        self.token_service.update_token_status(token.id, "FORBIDDEN")
                        return False
                    elif status_code == 429:
                        self.token_service.update_token_status(token.id, "RATE_LIMITED")
                        return False
                    elif 500 <= status_code < 600:
                        self.token_service.update_token_status(token.id, "SERVER_ERROR")
                        return False
                    else:
                        self.token_service.update_token_status(token.id, "UNKNOWN_ERROR")
                        return False

        except httpx.TimeoutException:
            logger.error("API请求超时")
            # 超时时设置为服务器错误状态
            self.token_service.update_token_status(token.id, "SERVER_ERROR")
            return False
        except Exception as e:
            logger.error(f"调用外部API失败: {e}")
            # 其他异常设置为未知错误状态
            self.token_service.update_token_status(token.id, "UNKNOWN_ERROR")
            return False
    
    def _get_portal_info(self, token: Token) -> Dict[str, Any]:
        """获取Token的门户信息"""
        try:
            # 从portal_url中提取token参数
            import urllib.parse as urlparse
            parsed_url = urlparse.urlparse(token.portal_url)
            query_params = urlparse.parse_qs(parsed_url.query)
            portal_token = query_params.get('token', [None])[0]

            if not portal_token:
                return {
                    "error": "portal_url中没有找到token参数",
                    "last_updated": "2024-01-01T00:00:00Z",
                    "status": "error"
                }

            # 第一步：获取客户信息
            customer_info = self._get_customer_from_link(portal_token)
            if not customer_info:
                return {
                    "error": "获取客户信息失败",
                    "last_updated": "2024-01-01T00:00:00Z",
                    "status": "error"
                }

            # 第二步：获取账户余额信息
            ledger_info = self._get_ledger_summary(customer_info, portal_token)
            if not ledger_info:
                return {
                    "error": "获取账户余额信息失败",
                    "last_updated": "2024-01-01T00:00:00Z",
                    "status": "error"
                }

            # 构建portal_info
            credits_balance_str = ledger_info.get("credits_balance", "0")
            try:
                # 处理可能的浮点数字符串，转换为整数
                credits_balance = int(float(credits_balance_str))
            except (ValueError, TypeError):
                credits_balance = 0

            portal_info = {
                "credits_balance": credits_balance,
                "is_active": False,
                "expiry_date": "",
                "last_updated": "2024-01-01T00:00:00Z",
                "status": "active"
            }

            # 设置is_active和expiry_date
            credit_blocks = ledger_info.get("credit_blocks", [])
            if credit_blocks:
                first_block = credit_blocks[0]
                portal_info["is_active"] = first_block.get("is_active", False)
                portal_info["expiry_date"] = first_block.get("expiry_date", "")

            # 如果余额为0，检查是否有无限制使用权限
            credits_balance = portal_info["credits_balance"]
            if credits_balance == 0:
                has_unlimited = self.check_subscription_info(token)
                portal_info["subscription_info"] = {
                    "plan_type": "unlimited" if has_unlimited else "limited"
                }

            return portal_info

        except Exception as e:
            logger.error(f"获取门户信息失败: {e}")
            return {
                "error": str(e),
                "last_updated": "2024-01-01T00:00:00Z",
                "status": "error"
            }

    def _get_customer_from_link(self, portal_token: str) -> Dict[str, Any]:
        """第一步：获取客户信息"""
        try:
            url = f"https://portal.withorb.com/api/v1/customer_from_link?token={portal_token}"
            headers = {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
                "Accept": "application/json, text/plain, */*",
                "Accept-Language": "en-US,en;q=0.9",
                "Accept-Charset": "utf-8",
                "Connection": "keep-alive",
                "Sec-Fetch-Dest": "empty",
                "Sec-Fetch-Mode": "cors",
                "Sec-Fetch-Site": "same-origin"
            }

            with httpx.Client(timeout=30.0) as client:
                response = client.get(url, headers=headers)

                if response.status_code == 200:
                    return response.json()
                else:
                    logger.warning(f"获取客户信息失败，状态码: {response.status_code}")
                    return None

        except Exception as e:
            logger.error(f"获取客户信息失败: {e}")
            return None

    def _get_ledger_summary(self, customer_info: Dict[str, Any], portal_token: str) -> Dict[str, Any]:
        """第二步：获取账户余额信息"""
        try:
            customer = customer_info.get("customer", {})
            customer_id = customer.get("id")
            ledger_pricing_units = customer.get("ledger_pricing_units", [])

            if not customer_id or not ledger_pricing_units:
                logger.error("客户信息中缺少必要字段")
                return None

            pricing_unit_id = ledger_pricing_units[0].get("id")
            if not pricing_unit_id:
                logger.error("pricing_unit_id不存在")
                return None

            url = f"https://portal.withorb.com/api/v1/customers/{customer_id}/ledger_summary?pricing_unit_id={pricing_unit_id}&token={portal_token}"
            headers = {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
                "Accept": "application/json, text/plain, */*",
                "Accept-Language": "en-US,en;q=0.9",
                "Accept-Charset": "utf-8",
                "Connection": "keep-alive",
                "Sec-Fetch-Dest": "empty",
                "Sec-Fetch-Mode": "cors",
                "Sec-Fetch-Site": "same-origin"
            }

            with httpx.Client(timeout=30.0) as client:
                response = client.get(url, headers=headers)

                if response.status_code == 200:
                    return response.json()
                else:
                    logger.warning(f"获取账户余额信息失败，状态码: {response.status_code}")
                    return None

        except Exception as e:
            logger.error(f"获取账户余额信息失败: {e}")
            return None

    def check_subscription_info(self, token: Token) -> bool:
        """检查订阅信息，判断是否有无限制使用权限"""
        try:
            # 构建API URL
            base_url = token.tenant_url.rstrip('/')
            url = f"{base_url}/subscription-info"

            # 设置请求头
            headers = {
                "Authorization": f"Bearer {token.access_token}",
                "Content-Type": "application/json"
            }

            # 发送请求
            with httpx.Client(timeout=30.0) as client:
                response = client.post(url, json={}, headers=headers)

                if response.status_code == 200:
                    response_text = response.text
                    # 检查响应中是否包含使用限制信息
                    has_usage_limit = "out of user messages" in response_text
                    return not has_usage_limit  # 如果包含限制信息则返回False，否则返回True
                else:
                    logger.warning(f"订阅信息检查失败，状态码: {response.status_code}")
                    return False

        except Exception as e:
            logger.error(f"检查订阅信息时发生异常: {str(e)}")
            return False
