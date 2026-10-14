import { memo, useEffect, useMemo, useState } from "react";
import { getCurrentApp } from "../../../../Api/helpers/helpers";
import { Button } from "../../../UI/Button";
import { Filters } from "../../../UI/Filters/Filters";
import cn from "classnames";
import styles from "./Styles.module.scss";
import { Checkbox } from "../../../UI/Checkbox";

type ScenarioStatus = "draft" | "published";

type Scenario = {
  id: string;
  title: string;
  status: ScenarioStatus;
  updated_at: string;
  enabled: boolean;
  hints?: { id: string }[];
};

const scenariosMock: Scenario[] = [
  {
    id: "1",
    title: "Первый сценарий",
    status: "draft",
    updated_at: "2026-09-28T10:00:00Z",
    enabled: false,
    hints: [{ id: "h1" }],
  },
  {
    id: "2",
    title: "Второй сценарий",
    status: "published",
    updated_at: "2026-10-08T12:00:00Z",
    enabled: true,
    hints: [
      { id: "h1" },
      { id: "h2" },
      { id: "h3" },
      { id: "h4" },
    ],
  },
];

function getTourStatus(item: Scenario) {
  return item.status === "published"
    ? {
        type: "published" as const,
        label: "Опубликован",
      }
    : {
        type: "draft" as const,
        label: "Черновик",
      };
}

function formatDate(date: string) {
  return new Date(date).toLocaleDateString("ru-RU");
}

function getStepsLabel(count: number) {
  if (count === 1) return "1 шаг";
  if (count >= 2 && count <= 4) return `${count} шага`;
  return `${count} шагов`;
}

function ScenariosComponent() {
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const [scenarios, setScenarios] = useState<Scenario[]>(scenariosMock);

  useEffect(() => {
    getCurrentApp()
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Не удалось загрузить приложение");
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  const filteredScenarios = useMemo(() => {
    return scenarios.filter((item) => {
      const matchesSearch = item.title.toLowerCase().includes(search.trim().toLowerCase());
      const matchesStatus = status === "all" || item.status === status;
      return matchesSearch && matchesStatus;
    });
  }, [scenarios, search, status]);

  const handleEdit = (id: string) => {
    console.log("edit", id);
  };

  const handleToggle = (id: string) => {
    setScenarios((prev) => prev.map((item) => item.id === id ? {...item, enabled: !item.enabled } : item));
  }

  if (isLoading) return <div>Загрузка...</div>;
  if (error) return <div>{error}</div>;

  return (
    <main className={styles.main}>
      <div className={styles.main__info}>
        <div className={styles.main__info__block}>
          <h1 className={styles.main__info__block__title}>Сценарии онбординга</h1>
          <p className={styles.main__info__block__p}>Список и подсказки</p>
        </div>
        <Button color="primary" className={styles.main__info__button}>+ Создать сценарий</Button>
      </div>

      <Filters
        search={search}
        status={status}
        onSearchChange={setSearch}
        onStatusChange={setStatus}
      />

      <table className={styles.page__table}>
        <thead>
          <tr className={styles.page__table__header}>
            <th>Название сценария</th>
            <th>Статус</th>
            <th>Обновлен</th>
            <th>Шагов</th>
            <th>Показ</th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody>
          {filteredScenarios.length === 0 ? (
            <tr>
              <td colSpan={6} className={styles.page__table__empty}>Сценарии не найдены</td>
            </tr>
          ) : (
            filteredScenarios.map((item) => {
              const tourStatus = getTourStatus(item);
              return (
                <tr className={styles.page__table__row} key={item.id}>
                  <td className={styles.page__table__row__name}>{item.title}</td>

                  <td>
                    <span className={cn(styles.page__table__row__status, tourStatus.type === "published"
                      ? styles.page__table__row__status__success : styles.page__table__row__status__warning)}
                    >
                      <span className={styles.page__table__row__status__dot} />
                      {tourStatus.label}
                    </span>
                  </td>

                  <td>{formatDate(item.updated_at)}</td>
                  <td>{getStepsLabel(item.hints?.length ?? 0)}</td>

                  <td className={styles.page__table__row__show}>
                    <Checkbox
                      checked={item.enabled}
                      onChange={() => handleToggle(item.id)}
                      label={item.enabled ? "Вкл" : "Выкл"}
                    />
                  </td>
                  
                  <td>
                    <Button 
                      size="edit" 
                      color="edit" 
                      className={styles.page__table__row__edit} 
                      onClick={() => handleEdit(item.id)}
                    >
                      Редактировать
                    </Button>
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </main>
  );
}

export const Scenarios = memo(ScenariosComponent);