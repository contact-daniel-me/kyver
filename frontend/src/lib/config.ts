const normalizeUrl = (value: string | undefined, fallback: string): string =>
  value && value.length > 0 ? value.replace(/\/$/, "") : fallback;

export const appConfig = {
  appName: process.env.NEXT_PUBLIC_APP_NAME ?? "Kyver",
  backendUrl: normalizeUrl(process.env.NEXT_PUBLIC_BACKEND_URL, "http://localhost:8000"),
};
