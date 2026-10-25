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
          editingDraftIndex, editingHintId, stepsCount, stepForm, setStepForm, submitHint, saveHint, stepErrors, setStepErrors,
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
                <div className={styles.fieldHeader}>
                  <span>Название шага</span>
                  {stepErrors.title && (<span className={styles.fieldError} role="alert">{stepErrors.title}</span>)}
                </div>
                <input
                  value={stepForm.title}
                  className={stepErrors.title ? styles.inputError : ""}
                  aria-invalid={Boolean(stepErrors.title)}
                  onChange={(e) => {setStepForm((prev) => ({...prev, title: e.target.value}));
                    setStepErrors((prev) => ({ ...prev, title: undefined }));}}
                  placeholder="Создать первое объявление"
                />
              </label>
              <label>
                <div className={styles.fieldHeader}>
                  <span>Текст подсказки</span>
                  {stepErrors.content && (<span className={styles.fieldError} role="alert">{stepErrors.content}</span>)}
                </div>
                <textarea
                  value={stepForm.content}
                  onChange={(e) => {setStepForm((form) => ({...form, content: e.target.value}));
                    setStepErrors((prev) => ({ ...prev, content: undefined }));}}
                  placeholder="Для того, чтобы создать свое первое объявление нажмите сюда..."
                  className={stepErrors.content ? styles.inputError : ""}
                />
              </label>
              <div className={styles.selectorField}>
                <div className={styles.fieldHeader}>
                  <label htmlFor="hint-selector">Элемент сайта</label>
                  {stepErrors.selector && (<span className={styles.fieldError} role="alert">{stepErrors.selector}</span>)}
                </div>

                <div className={styles.selectorRow}>
                  <input
                    id="hint-selector"
                    type="text"
                    placeholder="[data-onboarding='create-ad']"
                    value={stepForm.selector}
                    className={stepErrors.selector ? styles.inputError : ""}
                    aria-invalid={Boolean(stepErrors.selector)}
                    onChange={(e) => {setStepForm((prev) => ({...prev, selector: e.target.value}));
                      setStepErrors((prev) => ({...prev, selector: undefined}));}}
                  />
                  <button type="button" className={styles.pickButton} onClick={() => setIsPicking(true)}>
                    Выбрать
                  </button>
                </div>
              </div>
              <label>
                <div className={styles.fieldHeader}>
                  <span>Расположение подсказки</span>
                  {stepErrors.placement && (<span className={styles.fieldError} role="alert">{stepErrors.placement}</span>)}
                </div>
                <Select
                  value={stepForm.placement}
                  onChange={(value) => {setStepForm((prev) => ({ ...prev, placement: value as HintPlacement }));
                    setStepErrors((prev) => ({...prev, placement: undefined}));}}
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
