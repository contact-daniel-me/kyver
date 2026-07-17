from fastapi import APIRouter

from app.schemas.system import SystemCapabilitiesResponse
from app.services.capabilities import CORE_CAPABILITIES

router = APIRouter(prefix="/system", tags=["system"])


@router.get("/capabilities", response_model=SystemCapabilitiesResponse)
async def capabilities() -> SystemCapabilitiesResponse:
    return SystemCapabilitiesResponse(
        platform="Kyver",
        architecture="modular-multi-agent",
        capabilities=CORE_CAPABILITIES,
    )
