import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { onboardingAPI } from "../Api/onboarding";
import { getCurrentApp } from "../Api/helpers/helpers";

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
      const tours = await onboardingAPI.getTours((await getCurrentApp()).id);
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

  const submitScenario = () => {
    if (!scenario.name.trim()) {
      setScenarioErrors({ name: "Введите название сценария"});
      return;
    }
      setScenarioErrors({});
      saveScenario.mutate();
  };

  const saveScenario = useMutation({
    mutationFn: async () => {
      if (!scenario.name.trim()) throw new Error("Укажите название сценария");

      if (isNew) {
        if (!scenario.targetPath.trim()) throw new Error("Укажите путь на сайте");

        let id = createdTourIdRef.current;
        if (!id) {
          const tour = await onboardingAPI.createTour((await getCurrentApp()).id, {
            title: scenario.name.trim(), description: scenario.description,
            enabled: scenario.enabled, priority: scenario.priority,
            target_path: scenario.targetPath.trim(),
          });
          id = tour.id;
          createdTourIdRef.current = id;
        } else {
          await onboardingAPI.updateTourMeta(id, {
            title: scenario.name.trim(), description: scenario.description,
            enabled: scenario.enabled, priority: scenario.priority,
          });
        }
        await uploadDraftHints(id);
        return { id };
      }
      await onboardingAPI.updateTourMeta(tourId!, {
        title: scenario.name.trim(), description: scenario.description,
        enabled: scenario.enabled, priority: scenario.priority,
      });
      return { id: tourId! };
    },
    onSuccess: async (tour) => {
      setSaveError("");
      await queryClient.invalidateQueries({ queryKey: ["tours"] });
      if (isNew) navigate(`/admin/scenarios/${tour.id}/edit`, { replace: true });
      else await queryClient.invalidateQueries({ queryKey: ["scenario-edit-tour", tourId] });
    },
    onError: (error: Error) => setSaveError(error.message || "Не удалось сохранить сценарий и шаги"),
  });

  return { navigate, isNew, tourId, scenario, setScenario, saveError, saveScenario, tourQuery, scenarioErrors, setScenarioErrors, 
    submitScenario };
}
