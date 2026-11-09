import { onboardingAPI } from "../onboarding";
import { ONBOARDING_APP_KEY } from "../../Config/env";
import type { App } from "../../Types";

export async function getCurrentApp(): Promise<App | null> {
  const apps = await onboardingAPI.getApps();
  if (apps.length === 0) return null;

  if (ONBOARDING_APP_KEY) {
    const match = apps.find((app) => app.public_key === ONBOARDING_APP_KEY);
    return match ?? null;
  }

  return apps[0];
}