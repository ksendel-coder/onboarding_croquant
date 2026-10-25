import type { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import type { HintPlacement } from "../../../../Types";
import { Select } from "../../../UI/Select";
import { DomPicker } from "./DomPicker";
import styles from "./ScenarioEditor.module.scss";

type Editor = ReturnType<typeof useScenarioEditor>;

export function ScenarioSteps({ editor }: { editor: Editor }) {
  const { isNew, draftHints, hints, busy, draftLocked, moveHint, editDraftHint, removeDraftHint, editHint, removeHint, 
          editingDraftIndex, editingHintId, stepsCount, stepForm, setStepForm, submitHint, formError, saveHint,
          reorderHints, isPicking, pick, setIsPicking } = editor;

  return (
        <div className={styles.columns}>
          <section className={styles.card}>
            <ol className={styles.steps}>
              {isNew &&
                draftHints.map((hint, index) => (
                  <li key={`draft-${index}`}>
                      {hint.title}
                      {hint.selector}
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
                      <button type="button" disabled={busy || draftLocked} onClick={() => editDraftHint(index)}>
                        Изменить
                      </button>
                      <button type="button" disabled={busy || draftLocked} onClick={() => removeDraftHint(index)}>
                        Удалить
                      </button>
                    </div>
                  </li>
                ))}
              {!isNew &&
                hints.map((hint, index) => (
                  <li key={hint.id}>
                      {hint.title}
                      {hint.selector}
                    <div className={styles.actions}>
                      <button disabled={busy || index === 0} onClick={() => moveHint(index, -1)}>
                        ↑
                      </button>
                      <button disabled={busy || index === hints.length - 1} onClick={() => moveHint(index, 1)}>
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
