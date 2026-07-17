import type { SystemCapabilitiesResponse } from "@/types/agent";

interface PlatformOverviewProps {
  data: SystemCapabilitiesResponse;
}

export function PlatformOverview({ data }: PlatformOverviewProps) {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-8 px-6 py-16">
      <section className="space-y-3">
        <h1 className="text-4xl font-semibold tracking-tight">{data.platform}</h1>
        <p className="text-lg text-slate-600 dark:text-slate-300">
          AI-powered software engineering platform foundation built for modular, enterprise-scale
          evolution.
        </p>
        <p className="text-sm text-slate-500 dark:text-slate-400">Architecture: {data.architecture}</p>
      </section>

      <section className="grid gap-4 sm:grid-cols-2">
        {data.capabilities.map((capability) => (
          <article key={capability.name} className="rounded-lg border border-slate-200 p-4 dark:border-slate-800">
            <h2 className="font-medium">{capability.name}</h2>
            <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">{capability.description}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
