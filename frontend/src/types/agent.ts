export interface AgentCapability {
  name: string;
  description: string;
}

export interface SystemCapabilitiesResponse {
  platform: string;
  architecture: string;
  capabilities: AgentCapability[];
}
