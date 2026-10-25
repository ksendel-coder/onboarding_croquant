import { memo } from "react";
import { Button } from "../../../UI/Button";
import { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import { ScenarioSettings } from "./ScenarioSettings";
import { ScenarioSteps } from "./ScenarioSteps";
import styles from "./ScenarioEditor.module.scss";

function ScenarioEditorComponent() {
  const editor = useScenarioEditor();
  const { isNew, tourQuery, hintsQuery, navigate, busy, saveScenario, saveError, submitScenario } = editor;

  if (!isNew && (tourQuery.isLoading || hintsQuery.isLoading)) return <p>Загрузка сценария...</p>;
  if (!isNew && (tourQuery.isError || hintsQuery.isError || !tourQuery.data)) return <p>Не удалось загрузить сценарий.</p>;

  return (
    <main className={styles.page}>
      <button className={styles.back} type="button" onClick={() => navigate("/admin/scenarios")}>
        ← Назад к сценариям
      </button>
      <header className={styles.header}>
        <h1>{isNew ? "Создание сценария" : "Редактирование сценария"}</h1>
        <div className={styles.header__buttons}>
          <Button size="back" color="back" onClick={() => navigate("/admin/scenarios")}>
            Отмена
          </Button>
          <Button size="grey" color="grey" disabled={busy} onClick={submitScenario}>
            {saveScenario.isPending ? "Сохранение..." : "Сохранить черновик"}
          </Button>
           <Button size="pink" color="pink" disabled={true}>
            Опубликовать
          </Button>
        </div>
      </header>

      {saveError && <p role="alert">{saveError}</p>}
      <section className={styles.block}>
        <ScenarioSettings editor={editor} />
        <ScenarioSteps editor={editor} />
      </section>
    </main>
  );
}
export const ScenarioEditor = memo(ScenarioEditorComponent);
