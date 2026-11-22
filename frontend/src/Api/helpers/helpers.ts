import { onboardingAPI } from "../onboarding";
import { ONBOARDING_APP_KEY } from "../../Config/env";
import type { App } from "../../Types";

export async function getCurrentApp(): Promise<App | null> {
  return selectCurrentApp(await onboardingAPI.getApps());
}

export function selectCurrentApp(apps: App[]): App | null {
  return apps.find((app) => app.public_key === ONBOARDING_APP_KEY) ?? apps[0] ?? null;
}

export async function getOrCreateCurrentApp(): Promise<App> {
  const app = await getCurrentApp();
  if (app) return app;

  return onboardingAPI.createApp({
    name: "onboarding-demo",
    allowed_origins: [window.location.origin],
  });
}
