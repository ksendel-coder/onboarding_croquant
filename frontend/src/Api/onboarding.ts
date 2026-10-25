import { api } from "./api";
import type { App, CreateHintRequest, CreateTourRequest, Hint, HintPatchRequest, TourListItem, UpdateTourMetaRequest } from "../Types";

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

  createTour: (appId: string, body: CreateTourRequest) =>
    api.post<TourListItem>("/tours", body, { params: { appId } }).then((r) => r.data),
  

  getHints: (tourId: string) => api.get<Hint[]>(`/tours/${tourId}/hints`).then((r) => r.data),

  createHint: (tourId: string, body: CreateHintRequest) => api.post<Hint>(`/tours/${tourId}/hints`, body).then((r) => r.data),

  patchHint: (tourId: string, hintId: string, body: HintPatchRequest) =>
    api.patch<Hint>(`/tours/${tourId}/hints/${hintId}`, body).then((r) => r.data),

  deleteHint: (tourId: string, hintId: string) => api.delete(`/tours/${tourId}/hints/${hintId}`),

  reorderHints: (tourId: string, hintIds: string[]) =>
    api.put<Hint[]>(`/tours/${tourId}/hints/order`, { hint_ids: hintIds }).then((r) => r.data),
};