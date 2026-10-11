import { api } from "./api";
import type { App } from "../Types";

export const onboardingAPI = {
  getApps: () => api.get<App[]>("/apps").then((res) => res.data),

  createApp: (
    data: {
      name: string;
      allowed_origins: string[];
    },
  ) => api.post<App>("/apps", data).then((res) => res.data),
};