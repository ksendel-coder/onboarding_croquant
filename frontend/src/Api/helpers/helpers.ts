import { onboardingAPI } from "../onboarding";
import { ONBOARDING_APP_KEY } from "../../Config/env";
import type { App } from "../../Types";

export async function getCurrentApp(): Promise<App> {
  const apps = await onboardingAPI.getApps();

  if (ONBOARDING_APP_KEY) {
    const match = apps.find(
      (a) => a.public_key === ONBOARDING_APP_KEY,
    );

    if (!match) throw new Error("VITE_ONBOARDING_APP_KEY не совпадает ни с одним приложением");
    return match;
  }

  if (apps.length === 0) {
    const created = await onboardingAPI.createApp({
      name: "onboarding-demo",
      allowed_origins: [window.location.origin],
    });

    console.info("[Onboarding] created app public_key:", created.public_key);
    return created;
  }

  throw new Error("VITE_ONBOARDING_APP_KEY не задан, а в системе уже есть приложения");
}