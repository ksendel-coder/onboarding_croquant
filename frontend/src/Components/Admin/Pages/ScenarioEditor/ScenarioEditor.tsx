import { memo } from "react";
import { useIsMutating } from "@tanstack/react-query";
import { Button } from "../../../UI/Button";
import { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import { ScenarioSettings } from "./ScenarioSettings";
import { ScenarioSteps } from "./ScenarioSteps";
import { ThemeSettings } from "../ThemeSettings/ThemeSettings";
import styles from "./ScenarioEditor.module.scss";

function ScenarioEditorComponent() {
  const editor = useScenarioEditor();
  const savingTheme = useIsMutating({ mutationKey: ["save-app-theme"] }) > 0;
  const { isNew, tourQuery, hintsQuery, navigate, busy, saveScenario, saveError, submitScenario, publishScenario, 
          submitPublish } = editor;

  if (!isNew && (tourQuery.isLoading || hintsQuery.isLoading)) return <p>Загрузка сценария...</p>;
  if (!isNew && (tourQuery.isError || hintsQuery.isError || !tourQuery.data)) return <p>Не удалось загрузить сценарий.</p>;
  if (!isNew && editor.settingsQuery.isLoading) return <p role="status">Загрузка настроек сценария…</p>;
  if (!isNew && editor.settingsQuery.isError) {
    return (
      <main className={styles.page}>
        <button className={styles.back} type="button" onClick={() => navigate("/admin/scenarios")}>
          ← Назад к сценариям
        </button>
        <p role="alert">Не удалось загрузить настройки черновика.</p>
        <button type="button" onClick={() => void editor.settingsQuery.refetch()}>Повторить</button>
      </main>
    );
  }

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
          <Button size="grey" color="grey" disabled={busy || savingTheme} onClick={submitScenario}>
            {saveScenario.isPending ? "Сохранение..." : "Сохранить черновик"}
          </Button>
          <Button size="pink" color="pink" disabled={busy || savingTheme} onClick={submitPublish}>
            {publishScenario.isPending ? "Публикация..." : "Опубликовать"}
          </Button>
        </div>
      </header>

      {saveError && <p role="alert">{saveError}</p>}
      <section className={styles.block}>
        <ScenarioSettings editor={editor} />
        <ScenarioSteps editor={editor} targetPath={editor.settings.targetPath} />
      </section>
      <ThemeSettings key={editor.tourId ?? "new"} appId={tourQuery.data?.app_id} disabled={busy} />
    </main>
  );
}
export const ScenarioEditor = memo(ScenarioEditorComponent);
