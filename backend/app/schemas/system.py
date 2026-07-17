from pydantic import BaseModel


class CapabilitySet(BaseModel):
    name: str
    description: str


class SystemCapabilitiesResponse(BaseModel):
    platform: str
    architecture: str
    capabilities: list[CapabilitySet]
