from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
import urllib.parse
from typing import Optional

router = APIRouter()

class IDEOpenRequest(BaseModel):
    editor_type: str
    token: str
    tenant_url: str
    portal_url: Optional[str] = None

class JetBrainsTokenRequest(BaseModel):
    editor_type: str
    token: str
    tenant_url: str

@router.post("/open-editor")
async def open_editor(request: IDEOpenRequest):
    """在指定的IDE中打开Token"""
    try:
        # VSCode系列编辑器
        vscode_editors = {
            'vscode': 'vscode',
            'cursor': 'cursor', 
            'kiro': 'kiro',
            'trae': 'trae',
            'windsurf': 'windsurf',
            'qoder': 'qoder',
            'vscodium': 'vscodium',
            'codebuddy': 'codebuddy'
        }
        
        # JetBrains系列编辑器
        jetbrains_editors = {
            'idea': 'idea',
            'pycharm': 'pycharm',
            'goland': 'goland',
            'rustrover': 'rustrover',
            'webstorm': 'webstorm',
            'phpstorm': 'phpstorm',
            'androidstudio': 'androidstudio',
            'clion': 'clion',
            'datagrip': 'datagrip',
            'rider': 'rider',
            'rubymine': 'rubymine',
            'aqua': 'aqua'
        }
        
        if request.editor_type in vscode_editors:
            # 生成VSCode系列协议URL，返回给前端处理
            protocol_url = generate_vscode_protocol_url(
                request.editor_type,
                request.token,
                request.tenant_url,
                request.portal_url
            )

            return {
                "success": True,
                "message": f"正在打开 {request.editor_type}",
                "protocol_url": protocol_url
            }

        elif request.editor_type in jetbrains_editors:
            # 生成JetBrains编辑器协议URL，返回给前端处理
            protocol_url = generate_jetbrains_protocol_url(
                request.editor_type,
                request.token,
                request.tenant_url
            )

            return {
                "success": True,
                "message": f"正在打开 {request.editor_type}",
                "protocol_url": protocol_url
            }
            
        else:
            raise HTTPException(status_code=400, detail=f"不支持的编辑器类型: {request.editor_type}")
            
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"打开编辑器失败: {str(e)}")

def generate_vscode_protocol_url(editor_type: str, token: str, tenant_url: str, portal_url: Optional[str] = None) -> str:
    """生成VSCode系列编辑器的协议URL"""
    encoded_token = urllib.parse.quote(token)
    encoded_url = urllib.parse.quote(tenant_url)
    encoded_portal = urllib.parse.quote(portal_url or "")
    
    return f"{editor_type}://Augment.vscode-augment/autoAuth?token={encoded_token}&url={encoded_url}&portal={encoded_portal}"

def generate_jetbrains_protocol_url(editor_type: str, token: str, tenant_url: str) -> str:
    """生成JetBrains系列编辑器的协议URL"""
    encoded_token = urllib.parse.quote(token)
    encoded_url = urllib.parse.quote(tenant_url)
    
    return f"jetbrains://{editor_type}/plugin/Augment.jetbrains-augment/autoAuth?token={encoded_token}&url={encoded_url}"



@router.get("/supported-editors")
async def get_supported_editors():
    """获取支持的编辑器列表"""
    return {
        "vscode_editors": [
            {"id": "vscode", "name": "VS Code", "icon": "/icons/vscode.svg"},
            {"id": "cursor", "name": "Cursor", "icon": "/icons/cursor.svg"},
            {"id": "kiro", "name": "Kiro", "icon": "/icons/kiro.svg"},
            {"id": "trae", "name": "Trae", "icon": "/icons/trae.svg"},
            {"id": "windsurf", "name": "Windsurf", "icon": "/icons/windsurf.svg"},
            {"id": "qoder", "name": "Qoder", "icon": "/icons/qoder.svg"},
            {"id": "vscodium", "name": "VSCodium", "icon": "/icons/vscodium.svg"},
            {"id": "codebuddy", "name": "CodeBuddy", "icon": "/icons/codebuddy.svg"}
        ],
        "jetbrains_editors": [
            {"id": "idea", "name": "IntelliJ IDEA", "icon": "/icons/idea.svg"},
            {"id": "pycharm", "name": "PyCharm", "icon": "/icons/pycharm.svg"},
            {"id": "goland", "name": "GoLand", "icon": "/icons/goland.svg"},
            {"id": "rustrover", "name": "RustRover", "icon": "/icons/rustrover.svg"},
            {"id": "webstorm", "name": "WebStorm", "icon": "/icons/webstorm.svg"},
            {"id": "phpstorm", "name": "PhpStorm", "icon": "/icons/phpstorm.svg"},
            {"id": "androidstudio", "name": "Android Studio", "icon": "/icons/androidstudio.svg"},
            {"id": "clion", "name": "CLion", "icon": "/icons/clion.svg"},
            {"id": "datagrip", "name": "DataGrip", "icon": "/icons/datagrip.svg"},
            {"id": "rider", "name": "Rider", "icon": "/icons/rider.svg"},
            {"id": "rubymine", "name": "RubyMine", "icon": "/icons/rubymine.svg"},
            {"id": "aqua", "name": "Aqua", "icon": "/icons/aqua.svg"}
        ]
    }
