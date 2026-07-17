import { appConfig } from "@/lib/config";
import type { SystemCapabilitiesResponse } from "@/types/agent";

const FALLBACK_CAPABILITIES: SystemCapabilitiesResponse = {
  platform: "Kyver",
  architecture: "modular-multi-agent",
  capabilities: [],
};

export async function fetchSystemCapabilities(): Promise<SystemCapabilitiesResponse> {
  try {
    const response = await fetch(`${appConfig.backendUrl}/api/v1/system/capabilities`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return FALLBACK_CAPABILITIES;
    }

    return (await response.json()) as SystemCapabilitiesResponse;
  } catch {
    return FALLBACK_CAPABILITIES;
  }
}
