import { memo, useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../../../../Api/onboarding";
import { getCurrentApp } from "../../../../Api/helpers/helpers";
import type { CreateHintRequest, Hint, HintPlacement } from "../../../../Types";
import { isValidSelector } from "../../../../Utils/domSelector";
import { DomPicker } from "./DomPicker";
import styles from "./ScenarioEditor.module.scss";
import { Button } from "../../../UI/Button";
import { Select } from "../../../UI/Select";
import { Checkbox } from "../../../UI/Checkbox";

const emptyStep: CreateHintRequest = {
  title: "",
  content: "",
  selector: "",
  placement: "bottom",
};

const placements: HintPlacement[] = [
  "top",
  "bottom",
  "center",
  "left",
  "right",
];

function ScenarioEditorComponent() {
  const { tourId } = useParams<{ tourId?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isNew = !tourId;

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [targetPath, setTargetPath] = useState("/");
  const [enabled, setEnabled] = useState(false);
  const [priority, setPriority] = useState(0);
  const [editingHintId, setEditingHintId] = useState<string | null>(null);
  const [stepForm, setStepForm] = useState<CreateHintRequest>({ ...emptyStep });
  const [formError, setFormError] = useState("");
  const [isPicking, setIsPicking] = useState(false);
  const [draftHints, setDraftHints] = useState<CreateHintRequest[]>([]);
  const [editingDraftIndex, setEditingDraftIndex] = useState<number | null>(
    null,
  );
  const [saveError, setSaveError] = useState("");
  const createdTourIdRef = useRef<string | null>(null);
  const uploadedHintsCountRef = useRef(0);
  const [trigger, setTrigger] = useState("on_load");
  const [showOnce, setShowOnce] = useState(true);
  const [onlyNewUsers, setOnlyNewUsers] = useState(false);
  const [maxCount, setMaxCount] = useState("");
  const [placement, setPlacement] = useState("");

  const tourQuery = useQuery({
    queryKey: ["scenario-edit-tour", tourId],
    enabled: !isNew,
    queryFn: async () => {
      const app = await getCurrentApp();
      const tours = await onboardingAPI.getTours(app.id);
      return tours.find((tour) => tour.id === tourId) ?? null;
    },
  });

  const hintsQuery = useQuery({
    queryKey: ["hints", tourId],
    enabled: !isNew,
    queryFn: () => onboardingAPI.getHints(tourId!),
  });

  useEffect(() => {
    const tour = tourQuery.data;
    if (!tour) return;
    setName(tour.title);
    setDescription(tour.description ?? "");
    setEnabled(Boolean(tour.enabled));
    setPriority(tour.priority ?? 0);
  }, [tourQuery.data]);

  const saveScenario = useMutation({
    mutationFn: async () => {
      if (!name.trim()) throw new Error("Укажите название сценария");

      if (isNew) {
        if (!targetPath.trim()) throw new Error("Укажите путь на сайте");

        let id = createdTourIdRef.current;
        if (!id) {
          const app = await getCurrentApp();
          const tour = await onboardingAPI.createTour(app.id, {
            title: name.trim(),
            description,
            enabled,
            priority,
            target_path: targetPath.trim(),
          });
          id = tour.id;
          createdTourIdRef.current = id;
        } else {
          // Повторное нажатие после ошибки не создаёт второй сценарий.
          await onboardingAPI.updateTourMeta(id, {
            title: name.trim(),
            description,
            enabled,
            priority,
          });
        }

        for (
          let i = uploadedHintsCountRef.current;
          i < draftHints.length;
          i++
        ) {
          await onboardingAPI.createHint(id, draftHints[i]);
          uploadedHintsCountRef.current = i + 1;
        }
        return { id };
      }

      await onboardingAPI.updateTourMeta(tourId!, {
        title: name.trim(),
        description,
        enabled,
        priority,
      });
      return { id: tourId! };
    },
    onSuccess: async (tour) => {
      setSaveError("");
      await queryClient.invalidateQueries({ queryKey: ["tours"] });
      if (isNew) {
        navigate(`/admin/scenarios/${tour.id}/edit`, { replace: true });
      } else {
        await queryClient.invalidateQueries({
          queryKey: ["scenario-edit-tour", tourId],
        });
      }
    },
    onError: (error: Error) => {
      setSaveError(error.message || "Не удалось сохранить сценарий и шаги");
    },
  });

  const refreshHints = () =>
    queryClient.invalidateQueries({ queryKey: ["hints", tourId] });

  const saveHint = useMutation({
    mutationFn: (data: CreateHintRequest) => {
      if (!tourId) throw new Error("Сначала сохраните сценарий");
      if (!editingHintId) return onboardingAPI.createHint(tourId, data);
      const current = hintsQuery.data?.find(
        (hint) => hint.id === editingHintId,
      );
      if (!current) throw new Error("Шаг не найден");
      const patch: Partial<CreateHintRequest> = {};
      if (data.title !== current.title) patch.title = data.title;
      if (data.content !== current.content) patch.content = data.content;
      if (data.selector !== current.selector) patch.selector = data.selector;
      if (data.placement !== current.placement)
        patch.placement = data.placement;
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

  const pick = useCallback((selector: string) => {
    setStepForm((prev) => ({
      ...prev,
      selector,
    }));

    setIsPicking(false);
    setFormError("");
  }, []);

  const hints = [...(hintsQuery.data ?? [])].sort((a, b) => a.step - b.step);
  const stepsCount = isNew ? draftHints.length : hints.length;
  const busy =
    saveScenario.isPending ||
    saveHint.isPending ||
    removeHint.isPending ||
    reorderHints.isPending;
  const draftLocked = isNew && uploadedHintsCountRef.current > 0;

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
      setEditingDraftIndex((current) => {
        if (current === index) return destination;
        if (current === destination) return index;
        return current;
      });
      return;
    }

    const ids = hints.map((h) => h.id);
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
    setEditingDraftIndex((current) => {
      if (current === index) return null;
      if (current !== null && current > index) return current - 1;
      return current;
    });
    // Не удаляем отправленные шаги во время повторного сохранения.
  };

  const editHint = (hint: Hint) => {
    setEditingHintId(hint.id);
    setStepForm({
      title: hint.title,
      content: hint.content,
      selector: hint.selector,
      placement: hint.placement,
    });
    setFormError("");
  };

  const submitHint = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!stepForm.title.trim()) {
      setFormError("Введите название шага");
      return;
    }

    if (!stepForm.selector.trim() || !isValidSelector(stepForm.selector)) {
      setFormError("Введите корректный CSS-селектор");
      return;
    }

    setFormError("");

    if (isNew) {
      if (draftLocked) {
        setFormError(
          "Часть шагов уже отправлена. Повторите сохранение черновика.",
        );
        return;
      }
      if (editingDraftIndex === null) {
        setDraftHints((prev) => [...prev, { ...stepForm }]);
      } else {
        setDraftHints((prev) =>
          prev.map((hint, index) =>
            index === editingDraftIndex ? { ...stepForm } : hint,
          ),
        );
      }
      setEditingDraftIndex(null);
      setStepForm({ ...emptyStep });
      return;
    }

    saveHint.mutate(stepForm);
  };

  if (!isNew && (tourQuery.isLoading || hintsQuery.isLoading)) {
    return <p>Загрузка сценария...</p>;
  }
  if (!isNew && (tourQuery.isError || hintsQuery.isError || !tourQuery.data)) {
    return <p>Не удалось загрузить сценарий.</p>;
  }

  return (
    <main className={styles.page}>
      <button
        className={styles.back}
        type="button"
        onClick={() => navigate("/admin/scenarios")}
      >
        ← Назад к сценариям
      </button>
      <header className={styles.header}>
        <h1>{isNew ? "Создание сценария" : "Редактирование сценария"}</h1>
        <div className={styles.header__buttons}>
          <Button
            size="back"
            color="back"
            onClick={() => navigate("/admin/scenarios")}
          >
            Отмена
          </Button>
          <Button
            size="grey"
            color="grey"
            disabled={busy}
            onClick={() => saveScenario.mutate()}
          >
            {saveScenario.isPending ? "Сохранение..." : "Сохранить черновик"}
          </Button>
           <Button
            size="pink"
            color="pink"
            disabled={busy}
          >
            {saveScenario.isPending ? "Публикация..." : "Опубликовать"}
          </Button>
        </div>
      </header>

      {saveError && <p role="alert">{saveError}</p>}
      <section className={styles.block}>
        <section className={styles.card}>
          <form
            className={styles.fields}
            onSubmit={(event) => {
              event.preventDefault();
              saveScenario.mutate();
            }}
          >
            <label>
              Название сценария
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Например: Первое объявление"
              />
            </label>
            <label>
              Описание
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Зачем нужен этот сценарий?"
              />
            </label>
            {isNew && (
              <label>
                Путь на сайте
                <input
                  required
                  value={targetPath}
                  onChange={(e) => setTargetPath(e.target.value)}
                />
              </label>
            )}
            <label>
              Тип запуска
              <Select
                value={trigger}
                onChange={setTrigger}
                className={styles.select}
                options={[
                  { value: "on_load", label: "При загрузке" },
                  { value: "on_click", label: "По клику" },
                ]}
              />
            </label>
            <Checkbox
              checked={showOnce}
              onChange={setShowOnce}
              label="Показывать один раз"
            />

            <Checkbox
              checked={onlyNewUsers}
              onChange={setOnlyNewUsers}
              label="Только новым пользователям"
            />
            <label>
              Максимальное количество показов
              <input value={maxCount} onChange={(e) => setMaxCount(e.target.value)}/>
            </label>
          </form>
        </section>

        <div className={styles.columns}>
          <section className={styles.card}>
            <ol className={styles.steps}>
              {isNew &&
                draftHints.map((hint, index) => (
                  <li key={`draft-${index}`}>
                    <strong>
                      {index + 1}. {hint.title}
                    </strong>
                    <small>{hint.selector}</small>
                    <div className={styles.actions}>
                      <button
                        type="button"
                        disabled={busy || draftLocked || index === 0}
                        onClick={() => moveHint(index, -1)}
                      >
                        ↑
                      </button>
                      <button
                        type="button"
                        disabled={
                          busy || draftLocked || index === draftHints.length - 1
                        }
                        onClick={() => moveHint(index, 1)}
                      >
                        ↓
                      </button>
                      <button
                        type="button"
                        disabled={busy || draftLocked}
                        onClick={() => editDraftHint(index)}
                      >
                        Изменить
                      </button>
                      <button
                        type="button"
                        disabled={busy || draftLocked}
                        onClick={() => removeDraftHint(index)}
                      >
                        Удалить
                      </button>
                    </div>
                  </li>
                ))}
              {!isNew &&
                hints.map((hint, index) => (
                  <li key={hint.id}>
                    <strong>
                      {hint.step}. {hint.title}
                    </strong>
                    <small>{hint.selector}</small>
                    <div className={styles.actions}>
                      <button
                        disabled={busy || index === 0}
                        onClick={() => moveHint(index, -1)}
                      >
                        ↑
                      </button>
                      <button
                        disabled={busy || index === hints.length - 1}
                        onClick={() => moveHint(index, 1)}
                      >
                        ↓
                      </button>
                      <button disabled={busy} onClick={() => editHint(hint)}>
                        Изменить
                      </button>
                      <button
                        disabled={busy}
                        onClick={() => {
                          if (window.confirm(`Удалить шаг «${hint.title}»?`))
                            removeHint.mutate(hint.id);
                        }}
                      >
                        Удалить
                      </button>
                    </div>
                  </li>
                ))}
            </ol>
            <h2>Шаг {stepsCount + 1}</h2>
            <form className={styles.fields} onSubmit={submitHint}>
              <label>
                Название шага
                <input
                  value={stepForm.title}
                  onChange={(e) =>
                    setStepForm((form) => ({
                      ...form,
                      title: e.target.value,
                    }))
                  }
                  placeholder="Создать первое объявление"
                />
              </label>
              <label>
                Текст подсказки
                <textarea
                  value={stepForm.content}
                  onChange={(e) =>
                    setStepForm((form) => ({
                      ...form,
                      content: e.target.value,
                    }))
                  }
                  placeholder="Для того, чтобы создать свое первое объявление нажмите сюда..."
                />
              </label>
              <div className={styles.selectorField}>
                <label htmlFor="hint-selector">Элемент сайта</label>
                <div className={styles.selectorRow}>
                  <input
                    id="hint-selector"
                    type="text"
                    placeholder="[data-onboarding='create-ad']"
                    value={stepForm.selector}
                    onChange={(e) =>
                      setStepForm((prev) => ({
                        ...prev,
                        selector: e.target.value,
                      }))
                    }
                  />

                  <button
                    type="button"
                    className={styles.pickButton}
                    onClick={() => setIsPicking(true)}
                  >
                    Выбрать
                  </button>
                </div>
              </div>
              <label>
                Расположение подсказки
                <Select
                  value={placement}
                  onChange={setPlacement}
                  className={styles.select}
                  options={placements.map((p) => ({ value: p, label: p }))}
                />
              </label>
              {formError && <p role="alert">{formError}</p>}
              {saveHint.isError && (
                <p role="alert">Не удалось сохранить шаг.</p>
              )}
              {removeHint.isError && (
                <p role="alert">Не удалось удалить шаг.</p>
              )}
              {reorderHints.isError && (
                <p role="alert">Не удалось поменять порядок.</p>
              )}
              <button
                className={styles.primary}
                disabled={busy || draftLocked}
                type="submit"
              >
                {editingHintId !== null || editingDraftIndex !== null
                  ? "Сохранить изменения"
                  : "+ Добавить шаг"}
              </button>
            </form>
          </section>
          <DomPicker
            active={isPicking}
            onSelect={pick}
            onCancel={() => setIsPicking(false)}
          />
        </div>
      </section>
    </main>
  );
}

export const ScenarioEditor = memo(ScenarioEditorComponent)
