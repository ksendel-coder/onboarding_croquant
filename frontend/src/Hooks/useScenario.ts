import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { onboardingAPI } from "../Api/onboarding";
import { getCurrentApp, getOrCreateCurrentApp } from "../Api/helpers/helpers";
import {
  DEFAULT_TOUR_SETTINGS,
  validateAndSerializeTourSettings,
  type TourSettingsForm,
} from "../Types/tour";

export function useScenario(
  tourId: string | undefined,
  uploadDraftHints: (id: string) => Promise<void>,
) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isNew = !tourId;

  const [editedScenario, setEditedScenario] = useState<{
    name: string;
    description: string;
    enabled: boolean;
    priority: number;
  } | null>(null);
  const [editedSettings, setEditedSettings] = useState<TourSettingsForm | null>(null);
  const [saveError, setSaveError] = useState("");
  const [scenarioErrors, setScenarioErrors] = useState<{ name?: string }>({});
  const createdTourIdRef = useRef<string | null>(null);
  const submitting = useRef(false);

  const tourQuery = useQuery({
    queryKey: ["scenario-edit-tour", tourId],
    enabled: !isNew,
    queryFn: async () => {
      const app = await getCurrentApp();
      if (!app) return null;
      const tours = await onboardingAPI.getTours(app.id);
      return tours.find((tour) => tour.id === tourId) ?? null;
    },
  });

  // Preview возвращает настройки именно draft-версии, включая trigger_config.
  const settingsQuery = useQuery({
    queryKey: ["scenario-edit-settings", tourId],
    enabled: !isNew,
    retry: false,
    queryFn: () => onboardingAPI.getTourPreview(tourId!),
  });

  const scenario = editedScenario ?? {
    name: tourQuery.data?.title ?? "",
    description: tourQuery.data?.description ?? "",
    enabled: Boolean(tourQuery.data?.enabled),
    priority: tourQuery.data?.priority ?? 0,
  };

  const preview = settingsQuery.data;
  const settings: TourSettingsForm = editedSettings ?? {
    ...DEFAULT_TOUR_SETTINGS,
    targetPath: preview?.tour.target_path ?? tourQuery.data?.target_path ?? "/",
    triggerType: preview?.trigger_type === "element_visible" ? "on_load" : preview?.trigger_type ?? "on_load",
    delayMs: String(preview?.trigger_config?.delay_ms ?? 1000),
    scrollDepth: String(preview?.trigger_config?.scroll_depth ?? 50),
    inactivitySecs: String(preview?.trigger_config?.inactivity_secs ?? 30),
    showOnce: preview?.audience?.show_once ?? true,
    maxShows: String(preview?.audience?.max_shows ?? 0),
    onlyNew: preview?.audience?.only_new ?? false,
  };

  const setScenario = (updater: (prev: typeof scenario) => typeof scenario) => {
    setEditedScenario((prev) => updater(prev ?? scenario));
  };
  const setSettings = (updater: (prev: TourSettingsForm) => TourSettingsForm) => {
    setEditedSettings((prev) => updater(prev ?? settings));
  };

  const validateScenario = () => {
    if (!scenario.name.trim()) {
      setScenarioErrors({ name: "Введите название сценария" });
      return false;
    }
    setScenarioErrors({});
    try {
      validateAndSerializeTourSettings(settings);
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : "Неверные настройки сценария");
      return false;
    }
    return true;
  };

  const persistScenario = async (): Promise<{ id: string }> => {
    if (!scenario.name.trim()) throw new Error("Укажите название сценария");
    // Ошибки формы проверяем ДО POST /tours: не создавать частично неверный тур.
    const settingsBody = validateAndSerializeTourSettings(settings);
    if (!isNew && !settingsQuery.isSuccess) {
      throw new Error("Настройки черновика не загружены. Обновите страницу.");
    }

    const meta = {
      title: scenario.name.trim(),
      description: scenario.description,
      enabled: scenario.enabled,
      priority: scenario.priority,
    };

    if (isNew) {
      let id = createdTourIdRef.current;
      if (!id) {
        const app = await getOrCreateCurrentApp();
        if (!app) throw new Error("Сначала создайте приложение");
        const created = await onboardingAPI.createTour(app.id, {
          ...meta,
          target_path: settingsBody.target_path,
        });
        id = created.id;
        createdTourIdRef.current = id;
      } else {
        await onboardingAPI.updateTourMeta(id, meta);
      }
      await onboardingAPI.updateTourSettings(id, settingsBody);
      await uploadDraftHints(id);
      return { id };
    }

    await onboardingAPI.updateTourMeta(tourId!, meta);
    await onboardingAPI.updateTourSettings(tourId!, settingsBody);
    return { id: tourId! };
  };

  const afterSuccess = async () => {
    setSaveError("");
    const savedId = tourId ?? createdTourIdRef.current;
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["tours"] }),
      queryClient.invalidateQueries({ queryKey: ["scenario-edit-tour", savedId] }),
      queryClient.invalidateQueries({ queryKey: ["scenario-edit-settings", savedId] }),
      queryClient.invalidateQueries({ queryKey: ["tour-preview", savedId] }),
    ]);
    navigate("/admin/scenarios");
  };

  const saveScenario = useMutation({
    mutationFn: persistScenario,
    onSuccess: afterSuccess,
    onError: (error: Error) => setSaveError(error.message || "Не удалось сохранить сценарий"),
    onSettled: () => { submitting.current = false; },
  });

  const publishScenario = useMutation({
    mutationFn: async () => {
      const { id } = await persistScenario();
      await onboardingAPI.publishTour(id);
      return { id };
    },
    onSuccess: afterSuccess,
    onError: (error: Error) => setSaveError(error.message || "Не удалось опубликовать сценарий"),
    onSettled: () => { submitting.current = false; },
  });

  const submit = (mutate: () => void) => {
    if (submitting.current) return;
    setSaveError("");
    if (!validateScenario()) return;
    submitting.current = true;
    mutate();
  };

  return {
    navigate, isNew, tourId, scenario, setScenario, settings, setSettings,
    settingsQuery, saveError, saveScenario, publishScenario, tourQuery,
    scenarioErrors, setScenarioErrors,
    submitScenario: () => submit(saveScenario.mutate),
    submitPublish: () => submit(publishScenario.mutate),
  };
}
