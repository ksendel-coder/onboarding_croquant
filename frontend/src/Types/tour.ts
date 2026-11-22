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
  has_published: boolean;
  has_unpublished_changes: boolean;
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
  trigger_type: TourTriggerType;
  trigger_config: TourTriggerConfig | null;
  audience: {
    show_once: boolean;
    max_shows: number;
    only_new: boolean;
    rules: unknown[] | null;
  } | null;
}

export type TourTriggerType =
  | "on_load"
  | "delay"
  | "exit_intent"
  | "manual"
  | "scroll_depth"
  | "inactivity"
  | "element_visible";

export interface TourTriggerConfig {
  delay_ms?: number;
  scroll_depth?: number;
  inactivity_secs?: number;
  element_selector?: string;
}

export interface TourSettingsPatchRequest {
  target_path: string;
  trigger_type: TourTriggerType;
  trigger_config: TourTriggerConfig;
  audience_show_once: boolean;
  audience_max_shows: number;
  audience_only_new: boolean;
}

export interface TourSettingsForm {
  targetPath: string;
  triggerType: Exclude<TourTriggerType, "element_visible">;
  delaySecs: string;
  scrollDepth: string;
  inactivitySecs: string;
  showOnce: boolean;
  maxShows: string;
  onlyNew: boolean;
}

export const DEFAULT_TOUR_SETTINGS: TourSettingsForm = {
  targetPath: "/",
  triggerType: "on_load",
  delaySecs: "1",
  scrollDepth: "50",
  inactivitySecs: "30",
  showOnce: true,
  maxShows: "0",
  onlyNew: false,
};

export function validateAndSerializeTourSettings(
  form: TourSettingsForm,
): TourSettingsPatchRequest {
  const path = form.targetPath.trim();
  if (!path.startsWith("/") || path.startsWith("//") || /[?#\\\s]/.test(path)) {
    throw new Error("Путь должен начинаться с / и не содержать query/hash, например /widget-test");
  }

  const toInteger = (value: string, label: string, min: number, max: number) => {
    if (!/^\d+$/.test(value.trim())) throw new Error(`${label}: укажите целое число`);
    const number = Number(value);
    if (!Number.isSafeInteger(number) || number < min || number > max) {
      throw new Error(`${label}: допустимое значение от ${min} до ${max}`);
    }
    return number;
  };

  const config: TourTriggerConfig = {};
  if (form.triggerType === "delay") {
    const delay = form.delaySecs.trim().replace(",", ".");
    if (!/^(?:\d+(?:\.\d{1,3})?|\.\d{1,3})$/.test(delay) || Number(delay) > 600) {
      throw new Error("Задержка: укажите от 0 до 600 секунд, до трёх знаков после запятой");
    }
    config.delay_ms = Math.round(Number(delay) * 1000);
  }
  if (form.triggerType === "scroll_depth") {
    config.scroll_depth = toInteger(form.scrollDepth, "Прокрутка", 0, 100);
  }
  if (form.triggerType === "inactivity") {
    config.inactivity_secs = toInteger(form.inactivitySecs, "Бездействие", 0, 86400);
  }
  return {
    target_path: path,
    trigger_type: form.triggerType,
    trigger_config: config,
    audience_show_once: form.showOnce,
    audience_max_shows: toInteger(form.maxShows, "Максимум показов", 0, 2147483647),
    audience_only_new: form.onlyNew,
  };
}
