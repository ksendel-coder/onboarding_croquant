import type { Hint } from "./hint";

export type WidgetEventType = "tour_started" | "hint_shown" | "hint_completed" | "tour_completed" | "tour_dismissed";

export interface ResolvedTour {
  id: string;
  title: string;
  target_path: string;
  hints: Hint[];
  trigger_type?: string;
  trigger_config?: {
    delay_ms?: number;
    scroll_depth?: number;
    inactivity_secs?: number;
    element_selector?: string;
  };
}

export interface ResolveResult {
  tour: ResolvedTour;
  tour_version_id: string;
  version: number;
  current_hint_id?: string | null;
}

export interface WidgetEvent {
  event_id: string;
  attempt_id: string;
  subject_id: string;
  tour_id: string;
  tour_version_id: string;
  type: WidgetEventType;
  hint_id?: string;
}