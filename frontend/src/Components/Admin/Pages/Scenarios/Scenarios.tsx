import { memo, useMemo, useState } from "react";
import { Button } from "../../../UI/Button";
import { Filters } from "../../../UI/Filters/Filters";
import cn from "classnames";
import styles from "./Styles.module.scss";
import { Checkbox } from "../../../UI/Checkbox";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
// import { useDeleteTourMutation } from "../../../../Hooks/useDeleteTourMutation";
import { onboardingAPI } from "../../../../Api/onboarding";
import { useNavigate } from "react-router-dom";

function ScenariosComponent() {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const { data: scenarios = [], isLoading, error, refetch } = useToursQuery();
  // const deleteMutation = useDeleteTourMutation();
  const navigate = useNavigate();

  const filteredScenarios = useMemo(() => {
    return scenarios.filter((item) => {
      const matchesSearch = item.title.toLowerCase().includes(search.trim().toLowerCase());
      const matchesStatus = status === "all" || item.status === status;
      return matchesSearch && matchesStatus;
    });
  }, [scenarios, search, status]);

  // const handleDelete = (id: string) => { deleteMutation.mutate(id); };
  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    try {
      await onboardingAPI.updateTourMeta(id, { enabled });
      await refetch();
      
    } catch (error) {
      console.error("TOGGLE ENABLED ERROR", error);
    }
  };

  const handleEdit = (id: string) => {
    navigate(`/admin/scenarios/${id}/edit`);
  };

  if (isLoading) return <div>Загрузка...</div>;
  if (error) return <div>Ошибка загрузки сценариев</div>;

  return (
    <main className={styles.main}>
      <div className={styles.main__info}>
        <div className={styles.main__info__block}>
          <h1 className={styles.main__info__block__title}>Сценарии онбординга</h1>
          <p className={styles.main__info__block__p}>Список и подсказки</p>
        </div>
        <Button 
          color="primary" 
          className={styles.main__info__button} 
          onClick={() => navigate("/admin/scenarios/new")}>
          + Создать сценарий
        </Button>
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
          {filteredScenarios.map((item) => (
            <tr className={styles.page__table__row} key={item.id}>
              <td className={styles.page__table__row__name}>{item.title}</td>

              <td>
                <span className={cn(styles.page__table__row__status, item.status === "published"
                  ? styles.page__table__row__status__success : styles.page__table__row__status__warning)}
                >
                <span className={styles.page__table__row__status__dot} />
                  {item.status === "published" ? "Опубликован" : "Черновик"}
                </span>
              </td>

              <td>{item.updated_at ? new Date(item.updated_at).toLocaleDateString() : "—"}</td>
              <td>{item.steps_count ?? "—"}</td>

              <td className={styles.page__table__row__show}>
                <div className={styles.checkboxWrapper}>
                  <Checkbox
                    checked={Boolean(item.enabled)}
                    onChange={(checked) => handleToggleEnabled(item.id, checked)}
                    label={item.enabled ? "Вкл" : "Выкл"}
                  />
                </div>
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
          ))}
        </tbody>
      </table>
    </main>
  );
}

export const Scenarios = memo(ScenariosComponent);