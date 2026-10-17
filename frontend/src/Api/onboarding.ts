import { api } from "./api";
import type { App, TourListItem, UpdateTourMetaRequest } from "../Types";

export const onboardingAPI = {
  getApps: () => api.get<App[]>("/apps").then((res) => res.data),

  createApp: (
    data: {
      name: string;
      allowed_origins: string[];
    },
  ) => api.post<App>("/apps", data).then((res) => res.data),

  getTours: (appId: string) => api.get<TourListItem[]>("/tours", { params: { appId } }).then((res) => res.data),

  updateTourMeta: (tourId: string, data: UpdateTourMetaRequest) => api.patch(`/tours/${tourId}`, data).then((res) => res.data),

  deleteTour: (tourId: string) => api.delete(`/tours/${tourId}`),
};