import { useCallback, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../../../../Api/onboarding";
import { getCurrentApp } from "../../../../Api/helpers/helpers";
import type { CreateHintRequest, Hint, HintPlacement } from "../../../../Types";
import { isValidSelector } from "../../../../Utils/domSelector";
import { DomPicker } from "./DomPicker";
import "./ScenarioBuilder.module.scss";

const emptyStep: CreateHintRequest = {
  title: "",
  content: "",
  selector: "",
  placement: "bottom",
};

const placements: HintPlacement[] = [
  "top", "bottom", "center", "left-top", "right-top", "left", "right",
];

export function ScenarioBuilder() {
  const { tourId } = useParams<{ tourId: string }>();
  const navigate = useNavigate();
  const client = useQueryClient();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<CreateHintRequest>({ ...emptyStep });
  const [error, setError] = useState("");

  const hintsQuery = useQuery({
    queryKey: ["hints", tourId],
    queryFn: () => onboardingAPI.getHints(tourId!),
    enabled: Boolean(tourId),
  });

  const tourQuery = useQuery({
    queryKey: ["builder-tour", tourId],
    queryFn: async () => {
      const app = await getCurrentApp();
      const tours = await onboardingAPI.getTours(app.id);
      return tours.find((t) => t.id === tourId) ?? null;
    },
    enabled: Boolean(tourId),
  });

  const refresh = async () => {
    await client.invalidateQueries({ queryKey: ["hints", tourId] });
  };

  const save = useMutation({
    mutationFn: async (data: CreateHintRequest) => {
      if (!tourId) throw new Error("Не выбран сценарий");
      if (editingId) {
        // Передаём только изменившиеся поля, настоящий partial PATCH.
        const current = hintsQuery.data?.find((h) => h.id === editingId);
        if (!current) throw new Error("Шаг не найден");
        const patch: Partial<CreateHintRequest> = {};
        for (const key of ["title", "content", "selector", "placement"] as const) {
          if (data[key] !== current[key]) {
            // Обновляем значения через явные ветки ниже, без небезопасного any.
            if (key === "title") patch.title = data.title;
            if (key === "content") patch.content = data.content;
            if (key === "selector") patch.selector = data.selector;
            if (key === "placement") patch.placement = data.placement;
          }
        }
        return onboardingAPI.patchHint(tourId, editingId, patch);
      }
      return onboardingAPI.createHint(tourId, data);
    },
    onSuccess: async () => {
      setEditingId(null);
      setForm({ ...emptyStep });
      setError("");
      await refresh();
    },
  });

  const remove = useMutation({
    mutationFn: (hintId: string) => onboardingAPI.deleteHint(tourId!, hintId),
    onSuccess: refresh,
  });

  const reorder = useMutation({
    mutationFn: (ids: string[]) => onboardingAPI.reorderHints(tourId!, ids),
    onSuccess: refresh,
  });

  const move = (hints: Hint[], index: number, delta: number) => {
    const destination = index + delta;
    if (destination < 0 || destination >= hints.length || reorder.isPending) return;
    const ids = hints.map((h) => h.id);
    [ids[index], ids[destination]] = [ids[destination], ids[index]];
    reorder.mutate(ids);
  };

  const pick = useCallback((selector: string) => {
    setForm((f) => ({ ...f, selector }));
  }, []);

  const startEditing = (h: Hint) => {
    setEditingId(h.id);
    setForm({ title: h.title, content: h.content,
      selector: h.selector, placement: h.placement });
    setError("");
  };

  const submit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!form.selector.trim() || !isValidSelector(form.selector)) {
      setError("Введите корректный CSS-селектор"); return;
    }
    if (!form.title.trim()) {
      setError("Введите название шага"); return;
    }
    setError("");
    save.mutate(form);
  };

  if (!tourId) return <p>Не указан ID сценария.</p>;
  if (hintsQuery.isLoading || tourQuery.isLoading) return <p>Загрузка конструктора...</p>;
  if (hintsQuery.isError || tourQuery.isError) return <p>Не удалось загрузить сценарий.</p>;
  if (!tourQuery.data) return <p>Сценарий не найден в текущем приложении.</p>;

  const hints = [...(hintsQuery.data ?? [])].sort((a, b) => a.step - b.step);
  const busy = save.isPending || remove.isPending || reorder.isPending;

  return (
    <main className="scenarioBuilder">
      <header>
        <button type="button" onClick={() => navigate("/admin/scenarios")}>← К сценариям</button>
        <h1>{tourQuery.data.title}</h1>
        <p>Шаги черновика сценария</p>
      </header>
      <div className="builderColumns">
        <section>
          <h2>Шаги ({hints.length})</h2>
          <ol className="builderSteps">
            {hints.map((h, index) => (
              <li key={h.id}>
                <b>{h.step}. {h.title}</b>
                <small>{h.selector}</small>
                <div className="builderStepActions">
                  <button type="button" disabled={busy || index === 0}
                    onClick={() => move(hints, index, -1)}>↑</button>
                  <button type="button" disabled={busy || index === hints.length - 1}
                    onClick={() => move(hints, index, 1)}>↓</button>
                  <button type="button" disabled={busy} onClick={() => startEditing(h)}>Изменить</button>
                  <button type="button" disabled={busy} onClick={() => {
                    if (window.confirm(`Удалить шаг «${h.title}»?`)) remove.mutate(h.id);
                  }}>Удалить</button>
                </div>
              </li>
            ))}
          </ol>
          <button type="button" disabled={busy} onClick={() => {
            setEditingId(null); setForm({ ...emptyStep }); setError("");
          }}>+ Новый шаг</button>
          <h2>{editingId ? "Редактирование шага" : "Новый шаг"}</h2>
          <form className="builderForm" onSubmit={submit}>
            <label>Название шага
              <input value={form.title} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} />
            </label>
            <label>Текст подсказки
              <textarea value={form.content} onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))} />
            </label>
            <label>CSS-селектор
              <input value={form.selector} onChange={(e) => setForm((f) => ({ ...f, selector: e.target.value }))} placeholder="#start-button" />
            </label>
            <label>Расположение
              <select value={form.placement} onChange={(e) => setForm((f) => ({ ...f, placement: e.target.value as HintPlacement }))}>
                {placements.map((p) => <option value={p} key={p}>{p}</option>)}
              </select>
            </label>
            {error && <p role="alert">{error}</p>}
            {save.isError && <p role="alert">Ошибка сохранения. Проверьте backend.</p>}
            {remove.isError && <p role="alert">Не удалось удалить шаг.</p>}
            {reorder.isError && <p role="alert">Не удалось изменить порядок.</p>}
            <button disabled={busy} type="submit">{editingId ? "Сохранить изменения" : "Добавить шаг"}</button>
          </form>
        </section>
        <DomPicker onSelect={pick} />
      </div>
    </main>
  );
}