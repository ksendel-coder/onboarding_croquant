import type { Hint } from "./hint";

export interface TourListItem {
  id: string;
  app_id: string;
  title: string;
  description?: string;
  enabled: boolean;
  priority: number;
  updated_at: string;
  status?: "draft" | "published" | "archived";
  target_path?: string;
  steps_count?: number;
}

export interface UpdateTourMetaRequest {
  title?: string;
  description?: string;
  enabled?: boolean;
  priority?: number;
}

export interface CreateTourRequest {
  title: string;
  description: string;
  enabled: boolean;
  priority: number;
  target_path: string;
}

export interface TourPreviewResponse {
  preview: true;
  tour: {
    id: string;
    version_id: string;
    version: number;
    status: "draft";
    target_path: string;
    steps_count: number;
    hints: Hint[];
  };
  trigger_type: string;
  trigger_config: {
    scroll_depth?: number;
    inactivity_secs?: number;
    element_selector?: string;
    delay_ms?: number;
  } | null;
  audience: {
    show_once: boolean;
    max_shows: number;
    only_new: boolean;
    rules: unknown[] | null;
  } | null;
}