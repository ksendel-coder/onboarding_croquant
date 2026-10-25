import type { useScenarioEditor } from "../../../../Hooks/useScenarioEditor";
import { Select } from "../../../UI/Select";
import { Checkbox } from "../../../UI/Checkbox";
import styles from "./ScenarioEditor.module.scss";
import { useState } from "react";

type Editor = ReturnType<typeof useScenarioEditor>;

export function ScenarioSettings({ editor }: { editor: Editor }) {
    const [trigger, setTrigger] = useState("on_load");
    const [showOnce, setShowOnce] = useState(true);
    const [onlyNewUsers, setOnlyNewUsers] = useState(false);
    const [maxCount, setMaxCount] = useState("");
    const { scenario, setScenario, isNew, saveScenario } = editor;

  return (
        <section className={styles.card}>
          <form className={styles.fields} onSubmit={(event) => {event.preventDefault(); saveScenario.mutate();}}>
            <label>
              Название сценария
              <input
                required
                value={scenario.name}
                onChange={(e) => setScenario((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="Например: Первое объявление"
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
            {isNew && (
              <label>
                Путь на сайте
                <input
                  required
                  value={scenario.targetPath}
                  onChange={(e) => setScenario((prev) => ({ ...prev, targetPath: e.target.value }))}
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
  );
}
