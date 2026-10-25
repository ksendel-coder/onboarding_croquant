import type { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import type { HintPlacement } from "../../../../Types";
import { Select } from "../../../UI/Select";
import { DomPicker } from "./DomPicker";
import styles from "./ScenarioEditor.module.scss";
import { DndContext, KeyboardSensor, PointerSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { Pencil, Trash2 } from "lucide-react";
import { SortableStep } from "./SortableStep";


type Editor = ReturnType<typeof useScenarioEditor>;

export function ScenarioSteps({ editor }: { editor: Editor }) {
  const { isNew, draftHints, hints, busy, draftLocked, moveHint, editDraftHint, removeDraftHint, editHint, removeHint, 
          editingDraftIndex, editingHintId, stepsCount, stepForm, setStepForm, submitHint, formError, saveHint,
          reorderHints, isPicking, pick, setIsPicking } = editor;

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
                  useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }));

  const stepIds = isNew ? draftHints.map((_, index) => `draft-${index}`) : hints.map((hint) => hint.id);

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
  if (!over || active.id === over.id || busy || draftLocked) return;

  const oldIndex = stepIds.indexOf(String(active.id));
  const newIndex = stepIds.indexOf(String(over.id));
  if (oldIndex === -1 || newIndex === -1) return;
  moveHint(oldIndex, newIndex - oldIndex);
};

  return (
        <div className={styles.columns}>
          <section className={styles.card}>
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
              <SortableContext items={stepIds} strategy={verticalListSortingStrategy}>
                <ol className={styles.steps}>
                  {isNew ? draftHints.map((hint, index) => (
                    <SortableStep key={`draft-${index}`} id={`draft-${index}`} disabled={busy || draftLocked}>
                      <div className={styles.stepInfo}>
                        <strong>Шаг {index + 1} — {hint.title}</strong>
                        <small>{hint.selector}</small>
                      </div>

                      <div className={styles.stepActions}>
                        <button
                          type="button"
                          disabled={busy || draftLocked}
                          aria-label="Изменить шаг"
                          title="Изменить"
                          onClick={() => editDraftHint(index)}
                        >
                          <Pencil size={18} />
                        </button>
                        <button
                          type="button"
                          disabled={busy || draftLocked}
                          aria-label="Удалить шаг"
                          title="Удалить"
                          className={styles.deleteButton}
                          onClick={() => removeDraftHint(index)}
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </SortableStep>
                  ))
                    : hints.map((hint, index) => (
                      <SortableStep key={hint.id} id={hint.id} disabled={busy}>
                        <div className={styles.stepInfo}>
                          <strong>Шаг {index + 1} — {hint.title}</strong>
                          <small>{hint.selector}</small>
                        </div>

                        <div className={styles.stepActions}>
                          <button
                          type="button"
                          disabled={busy}
                          aria-label="Изменить шаг"
                          title="Изменить"
                          onClick={() => editHint(hint)}
                        >
                          <Pencil size={18} />
                        </button>
                        <button
                          type="button"
                          disabled={busy}
                          aria-label="Удалить шаг"
                          title="Удалить"
                          className={styles.deleteButton}
                          onClick={() => { if (window.confirm(`Удалить шаг «${hint.title}»?`)) { removeHint.mutate(hint.id); }}}
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </SortableStep>
                  ))}
                </ol>
              </SortableContext>
            </DndContext>
            <h2>
                Шаг {editingDraftIndex !== null ? editingDraftIndex + 1 : editingHintId !== null 
                ? (hints.find((h) => h.id === editingHintId)?.step ?? stepsCount + 1) : stepsCount + 1}
            </h2>
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

                  <button type="button" className={styles.pickButton} onClick={() => setIsPicking(true)}>
                    Выбрать
                  </button>
                </div>
              </div>
              <label>
                Расположение подсказки
                <Select
                  value={stepForm.placement}
                  onChange={(value) => setStepForm((prev) => ({ ...prev, placement: value as HintPlacement }))}
                  className={styles.select}
                  options={[
                    { value: "top", label: "Сверху" },
                    { value: "bottom", label: "Снизу" },
                    { value: "center", label: "По центру" },
                    { value: "left", label: "Слева" },
                    { value: "right", label: "Справа" },
                    { value: "left-top", label: "Сверху слева" },
                    { value: "right-top", label: "Сверху справа" },
                    ]}
                />
              </label>

              {formError && <p role="alert">{formError}</p>}
              {saveHint.isError && (<p role="alert">Не удалось сохранить шаг.</p>)}
              {removeHint.isError && (<p role="alert">Не удалось удалить шаг.</p>)}
              {reorderHints.isError && (<p role="alert">Не удалось поменять порядок.</p>)}

              <button className={styles.primary} disabled={busy || draftLocked} type="submit">
                {editingHintId !== null || editingDraftIndex !== null ? "Сохранить изменения" : "+ Добавить шаг"}
              </button>
            </form>
          </section>
          <DomPicker active={isPicking} onSelect={pick} onCancel={() => setIsPicking(false)} />
        </div>
  );
}
