import type { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import { Select } from "../../../UI/Select";
import { Checkbox } from "../../../UI/Checkbox";
import styles from "./ScenarioEditor.module.scss";
import type { TourTriggerType } from "../../../../Types/tour";

type Editor = ReturnType<typeof useScenarioEditor>;

export function ScenarioSettings({ editor }: { editor: Editor }) {
  const { scenario, setScenario, settings, setSettings } = editor;
  const change = <K extends keyof typeof settings>(key: K, value: (typeof settings)[K]) => {
    setSettings((previous) => ({ ...previous, [key]: value }));
  };

  return (
    <section className={styles.card}>
      <div className={styles.fields}>
        <label>
          <div className={styles.field__header}>
            <span>Название сценария</span>
            {editor.scenarioErrors.name && (
              <span className={styles.field__errorr} role="alert">
                {editor.scenarioErrors.name}
              </span>
            )}
          </div>
          <input
            value={scenario.name}
            className={editor.scenarioErrors.name ? styles.field__inputError : ""}
            aria-invalid={Boolean(editor.scenarioErrors.name)}
            placeholder="Например: Первое объявление"
            onChange={(e) => {
              setScenario((prev) => ({ ...prev, name: e.target.value }));
              editor.setScenarioErrors((prev) => ({ ...prev, name: undefined }));
            }}
          />
        </label>

        <label>
          Описание
          <textarea
            value={scenario.description}
            onChange={(e) => setScenario((prev) => ({ ...prev, description: e.target.value }))}
            placeholder="Зачем нужен этот сценарий?"
          />
        </label>

        <label>
          Путь на сайте
          <input
            required
            value={settings.targetPath}
            placeholder="/widget-test"
            onChange={(e) => change("targetPath", e.target.value)}
          />
        </label>

        <label>
          Тип запуска
          <Select
            value={settings.triggerType}
            onChange={(v) => change("triggerType", v as TourTriggerType)}
            className={styles.select}
            options={[
              { value: "on_load", label: "При загрузке страницы" },
              { value: "delay", label: "С задержкой" },
              { value: "scroll_depth", label: "После прокрутки" },
              { value: "inactivity", label: "После бездействия" },
              { value: "element_visible", label: "При появлении элемента" },
              { value: "exit_intent", label: "При попытке покинуть страницу" },
              { value: "manual", label: "Ручной запуск" },
            ]}
          />
        </label>
        {settings.triggerType !== "on_load" && (
          <small>Параметры сохраняются в сценарии. Поддержка этого триггера в виджете пока не подключена.</small>
        )}

        {settings.triggerType === "delay" && (
          <label>
            Задержка, мс (0–600000)
            <input type="number" min="0" max="600000"
              value={settings.delayMs}
              onChange={(e) => change("delayMs", e.target.value)} />
          </label>
        )}
        {settings.triggerType === "scroll_depth" && (
          <label>
            Прокрутка страницы, % (0–100)
            <input type="number" min="0" max="100"
              value={settings.scrollDepth}
              onChange={(e) => change("scrollDepth", e.target.value)} />
          </label>
        )}
        {settings.triggerType === "inactivity" && (
          <label>
            Время бездействия, сек. (0–86400)
            <input type="number" min="0" max="86400"
              value={settings.inactivitySecs}
              onChange={(e) => change("inactivitySecs", e.target.value)} />
          </label>
        )}
        {settings.triggerType === "element_visible" && (
          <label>
            CSS-селектор элемента запуска
            <input value={settings.elementSelector} placeholder="#start-button"
              onChange={(e) => change("elementSelector", e.target.value)} />
          </label>
        )}

        <Checkbox
          checked={settings.showOnce}
          onChange={(v) => change("showOnce", v)}
          label="Показывать один раз"
        />
        <Checkbox
          checked={settings.onlyNew}
          onChange={(v) => change("onlyNew", v)}
          label="Только новым пользователям"
        />
        <label>
          Максимальное количество показов (0 — без ограничения)
          <input type="number" min="0" step="1" value={settings.maxShows}
            onChange={(e) => change("maxShows", e.target.value)} />
        </label>
      </div>
    </section>
  );
}
