import { onboardingAPI } from "../onboarding";
import { ONBOARDING_APP_KEY } from "../../Config/env";
import type { App } from "../../Types";

export async function getCurrentApp(): Promise<App | null> {
  const apps = await onboardingAPI.getApps();
  if (apps.length === 0) return null;
  
  if (ONBOARDING_APP_KEY) return ( apps.find((app) => app.public_key === ONBOARDING_APP_KEY) ?? apps[0]);
  
  return apps[0];
}

export async function getOrCreateCurrentApp(): Promise<App> {
  const app = await getCurrentApp();
  if (app) return app;

  return onboardingAPI.createApp({
    name: "onboarding-demo",
    allowed_origins: [window.location.origin],
  });
}