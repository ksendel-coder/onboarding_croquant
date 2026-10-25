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