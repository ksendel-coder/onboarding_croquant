export interface TourVersionHistoryItem {
  id: string;
  version: number;
  status: "draft" | "published" | "archived";
  published_at?: string | null;
}

export interface StepAnalytics {
  hint_id: string;
  step: number;
  title?: string;
  shown: number;
  completed: number;
  skipped: number;
  drop_off: number;
  not_completed?: number;
  last_skipped_at?: string;
}
export interface AnalyticsPeriod {
  started: number;
  completed: number;
  dismissed: number;
}

export interface TourAnalytics {
  tour_id: string;
  tour_version_id: string;
  started: number;
  completed: number;
  dismissed: number;
  completion_rate: number; 
  steps: StepAnalytics[];
  recent?: AnalyticsPeriod;
  previous?: AnalyticsPeriod;
}
