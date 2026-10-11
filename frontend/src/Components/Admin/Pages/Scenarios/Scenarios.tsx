import { memo, useEffect, useMemo, useState } from "react";
import { getCurrentApp } from "../../../../Api/helpers/helpers";
import { Button } from "../../../UI/Button";
import { Filters } from "../../../UI/Filters/Filters";
import cn from "classnames";
import styles from "./Styles.module.scss";

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
    return scenariosMock.filter((item) => {
      const matchesSearch = item.title.toLowerCase().includes(search.trim().toLowerCase());
      const matchesStatus = status === "all" || item.status === status;
      return matchesSearch && matchesStatus;
    });
  }, [search, status]);

  const handleEdit = (id: string) => {
    console.log("edit", id);
  };

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

      <div className={styles.page__table}>
        <div className={styles.page__table__header}>
          <span>Название сценария</span>
          <span>Статус</span>
          <span>Обновлен</span>
          <span>Шагов</span>
          <span>Показ</span>
          <span>Действия</span>
        </div>

        {filteredScenarios.length === 0 ? (
          <div className={styles.page__table__empty}>Сценарии не найдены</div>
        ) : (
          filteredScenarios.map((item) => {
            const tourStatus = getTourStatus(item);
            return (
              <div className={styles.page__table__row} key={item.id}>
                <span className={styles.page__table__row__name}>{item.title}</span>

                <span>
                  <span
                    className={cn(styles.page__table__row__status, tourStatus.type === "published"
                      ? styles.page__table__row__status__success : styles.page__table__row__status__warning)}
                  >
                    <span className={styles.page__table__row__status__dot}/>
                    {tourStatus.label}
                  </span>
                </span>

                <span>{formatDate(item.updated_at)}</span>
                <span>{getStepsLabel(item.hints?.length ?? 0)}</span>

                <span className={styles.page__table__row__show}>
                  <span className={cn(styles.page__table__row__checkbox, item.enabled 
                    && styles.page__table__row__checkbox__checked)}
                  >
                    {item.enabled && "✓"}
                  </span>
                  <span>{item.enabled ? "Вкл" : "Выкл"}</span>
                </span>

                <span className={styles.page__table__row__actions}>
                  <button type="button" className={styles.page__table__row__edit} onClick={() => handleEdit(item.id)}>
                    Редактировать
                  </button>
                </span>
              </div>
            );
          })
        )}
      </div>
    </main>
  );
}

export const Scenarios = memo(ScenariosComponent);