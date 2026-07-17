import { PlatformOverview } from "@/features/platform/components/platform-overview";
import { fetchSystemCapabilities } from "@/lib/api/client";

export default async function Home() {
  const data = await fetchSystemCapabilities();

  return <PlatformOverview data={data} />;
}
