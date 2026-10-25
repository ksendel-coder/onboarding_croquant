import { useCallback, useRef, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import type { CreateHintRequest, Hint } from "../Types";
import { isValidSelector } from "../Utils/domSelector";

export const emptyStep: CreateHintRequest = { title: "", content: "", selector: "", placement: "bottom" };

export function useScenarioSteps(tourId: string | undefined) {
  const isNew = !tourId;
  const queryClient = useQueryClient();
  const [draftHints, setDraftHints] = useState<CreateHintRequest[]>([]);
  const [editingDraftIndex, setEditingDraftIndex] = useState<number | null>(null);
  const [editingHintId, setEditingHintId] = useState<string | null>(null);
  const [stepForm, setStepForm] = useState<CreateHintRequest>({ ...emptyStep });
  const [formError, setFormError] = useState("");
  const [isPicking, setIsPicking] = useState(false);
  const uploadedHintsCountRef = useRef(0);
  const [uploadStarted, setUploadStarted] = useState(false);

  const hintsQuery = useQuery({
    queryKey: ["hints", tourId],
    enabled: Boolean(tourId),
    queryFn: () => onboardingAPI.getHints(tourId!),
  });

  const refreshHints = () => queryClient.invalidateQueries({ queryKey: ["hints", tourId] });

  const saveHint = useMutation({
    mutationFn: (data: CreateHintRequest) => {
      if (!tourId) throw new Error("Сценарий пока не создан");
      if (!editingHintId) return onboardingAPI.createHint(tourId, data);

      const current = hintsQuery.data?.find((hint) => hint.id === editingHintId);
      if (!current) throw new Error("Шаг не найден");

      const patch: Partial<CreateHintRequest> = {};
      if (data.title !== current.title) patch.title = data.title;
      if (data.content !== current.content) patch.content = data.content;
      if (data.selector !== current.selector) patch.selector = data.selector;
      if (data.placement !== current.placement) patch.placement = data.placement;
      return onboardingAPI.patchHint(tourId, editingHintId, patch);
    },
    onSuccess: async () => {
      setEditingHintId(null);
      setStepForm({ ...emptyStep });
      setFormError("");
      await refreshHints();
    },
  });
  const removeHint = useMutation({
    mutationFn: (id: string) => onboardingAPI.deleteHint(tourId!, id),
    onSuccess: refreshHints,
  });
  const reorderHints = useMutation({
    mutationFn: (ids: string[]) => onboardingAPI.reorderHints(tourId!, ids),
    onSuccess: refreshHints,
  });

  const hints = [...(hintsQuery.data ?? [])].sort((a, b) => a.step - b.step);
  const stepsCount = isNew ? draftHints.length : hints.length;
  const busy = saveHint.isPending || removeHint.isPending || reorderHints.isPending;
  const draftLocked = isNew && uploadStarted;

  const uploadDraftHints = async (id: string) => {
    setUploadStarted(true);
    for (let index = uploadedHintsCountRef.current; index < draftHints.length; index++) {
      await onboardingAPI.createHint(id, draftHints[index]);
      uploadedHintsCountRef.current = index + 1;
    }
  };

  const pick = useCallback((selector: string) => {
    setStepForm((prev) => ({ ...prev, selector }));
    setIsPicking(false);
    setFormError("");
  }, []);

  const moveHint = (index: number, delta: number) => {
    const length = isNew ? draftHints.length : hints.length;
    const destination = index + delta;
    if (destination < 0 || destination >= length || busy || draftLocked) return;

    if (isNew) {
      setDraftHints((prev) => {
        const next = [...prev];
        [next[index], next[destination]] = [next[destination], next[index]];
        return next;
      });
      setEditingDraftIndex((current) => current === index ? destination : current === destination ? index : current);
      return;
    }
    const ids = hints.map((hint) => hint.id);
    [ids[index], ids[destination]] = [ids[destination], ids[index]];
    reorderHints.mutate(ids);
  };

  const editDraftHint = (index: number) => {
    if (draftLocked) return;
    setEditingDraftIndex(index);
    setEditingHintId(null);
    setStepForm({ ...draftHints[index] });
    setFormError("");
  };

  const removeDraftHint = (index: number) => {
    if (draftLocked) return;
    setDraftHints((prev) => prev.filter((_, i) => i !== index));
    setEditingDraftIndex((current) => current === index ? null : current !== null && current > index ? current - 1 : current);
  };

  const editHint = (hint: Hint) => {
    setEditingHintId(hint.id);
    setStepForm({ title: hint.title, content: hint.content, selector: hint.selector, placement: hint.placement });
    setFormError("");
  };

  const submitHint = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!stepForm.title.trim()) { setFormError("Введите название шага"); return; }
    if (!stepForm.selector.trim() || !isValidSelector(stepForm.selector)) {
      setFormError("Введите корректный CSS-селектор"); return;
    }
    setFormError("");

    if (isNew) {
      if (draftLocked) { setFormError("Сохранение начато. Повторите сохранение черновика."); return; }
      if (editingDraftIndex === null) setDraftHints((prev) => [...prev, { ...stepForm }]);
      else setDraftHints((prev) => prev.map((hint, index) => index === editingDraftIndex ? { ...stepForm } : hint));
      setEditingDraftIndex(null);
      setStepForm({ ...emptyStep });
      return;
    }
    saveHint.mutate(stepForm);
  };

  return { draftHints, editingDraftIndex, editingHintId, stepForm, setStepForm, formError, setFormError, isPicking, setIsPicking, 
    hintsQuery, hints, stepsCount, busy, draftLocked, saveHint, removeHint, reorderHints, uploadDraftHints, moveHint, 
    editDraftHint, removeDraftHint, editHint, submitHint, pick };
}
