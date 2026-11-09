import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { onboardingAPI } from "../Api/onboarding";
import { getCurrentApp, getOrCreateCurrentApp } from "../Api/helpers/helpers";

export function useScenario(tourId: string | undefined, uploadDraftHints: (id: string) => Promise<void>) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isNew = !tourId;
  const [editedScenario, setEditedScenario] = useState<{
      name: string;
      description: string;
      targetPath: string;
      enabled: boolean;
      priority: number;
  } | null>(null);

  const [saveError, setSaveError] = useState("");
  const createdTourIdRef = useRef<string | null>(null);

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

  const scenario = editedScenario ?? {
    name: tourQuery.data?.title ?? "",
    description: tourQuery.data?.description ?? "",
    targetPath: "/",
    enabled: Boolean(tourQuery.data?.enabled),
    priority: tourQuery.data?.priority ?? 0,
  };

  const [scenarioErrors, setScenarioErrors] = useState<{name?: string;}>({});

  const setScenario = (updater: (prev: typeof scenario) => typeof scenario) => {
    setEditedScenario((prev) => updater(prev ?? scenario));
  };

  const validateScenario = () => {
    if (!scenario.name.trim()) {
      setScenarioErrors({ name: "Введите название сценария" });
      return false;
    }

    setScenarioErrors({});
    return true;
  };

  const persistScenario = async (): Promise<{ id: string }> => {
    if (!scenario.name.trim()) throw new Error("Укажите название сценария");

    const data = {
      title: scenario.name.trim(),
      description: scenario.description,
      enabled: scenario.enabled,
      priority: scenario.priority,
    };

    if (isNew) {
      if (!scenario.targetPath.trim()) throw new Error("Укажите путь на сайте");
  
      let id = createdTourIdRef.current;
      if (!id) {
        const app = await getOrCreateCurrentApp();
        if (!app) throw new Error("Сначала создайте приложение");

        const tour = await onboardingAPI.createTour(app.id, {...data, target_path: scenario.targetPath.trim()});
        id = tour.id;
        createdTourIdRef.current = id;

      } else {
        await onboardingAPI.updateTourMeta(id, data);
      }
      await uploadDraftHints(id);
      return { id };
    }
    await onboardingAPI.updateTourMeta(tourId!, data);
    return { id: tourId! };
  };

  const saveScenario = useMutation({
    mutationFn: persistScenario,

    onSuccess: async () => {
      setSaveError("");
      await queryClient.invalidateQueries({ queryKey: ["tours"] });
      navigate("/admin/scenarios");
    },

    onError: (error: Error) => { setSaveError(error.message || "Не удалось сохранить сценарий"); }
  });

  const publishScenario = useMutation({
    mutationFn: async () => {
      const { id } = await persistScenario();
      await onboardingAPI.publishTour(id);
      return { id };
    },

    onSuccess: async ({ id }) => {
      setSaveError("");

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["tours"] }),
        queryClient.invalidateQueries({ queryKey: ["scenario-edit-tour", id] }),
        queryClient.invalidateQueries({ queryKey: ["hints", id] }),
      ]);

      navigate(`/admin/scenarios/${id}/edit`, { replace: true });
    },

    onError: (error: Error) => { setSaveError(error.message || "Не удалось опубликовать сценарий"); }
  });

  const submitScenario = () => {
    if (!validateScenario()) return;
    setSaveError("");
    saveScenario.mutate();
  };

  const submitPublish = () => {
    if (!validateScenario()) return;
    setSaveError("");
    publishScenario.mutate();
  };

  return { navigate, isNew, tourId, scenario, setScenario, saveError, saveScenario, publishScenario, tourQuery,
           scenarioErrors, setScenarioErrors, submitScenario, submitPublish };
}