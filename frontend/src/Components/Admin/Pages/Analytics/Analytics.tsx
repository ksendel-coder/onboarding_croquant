import { useNavigate } from "react-router-dom";
import styles from "./Analytics.module.scss";

const demoRows = [
  {
    id: "demo-tour-1",
    name: "Публикация",
    version: "v7",
    started: 24150,
    completed: 20527,
    completionRate: 85.0,
    goalRate: 70.0,
  },
];

const demoMetrics: { label: string; value: string; note: string; kind?: "success" | "brand" }[] = [
  { label: "ВСЕГО ЗАПУСКОВ", value: "42,800", note: "↑ +14.2%" },
  { label: "ЗАВЕРШИЛИ ТУР", value: "36,380", note: "↑ +18.0%" },
  { label: "COMPLETION RATE", value: "85.0%", note: "Выше среднего", kind: "success" },
  { label: "GOAL RATE (ЦЕЛЬ)", value: "68.4%", note: "↑ +4.1%", kind: "brand" },
];

export function Analytics() {
    const navigate = useNavigate();
  return (
    <main className={styles.page}>
      <header className={styles.page__header}>
        <h1>Дашборд конверсий онбординга</h1>
      </header>

      <section className={styles.page__metrics} aria-label="Показатели онбординга">
        {demoMetrics.map((metric) => (
          <article className={styles.page__metrics__metricCard} key={metric.label}>
            <span className={styles.page__metrics__metricCard__metricLabel}>{metric.label}</span>
            <strong className={metric.kind === "success" ? styles.page__metrics__success : metric.kind === "brand" 
                ? styles.page__metrics__brand : undefined}>
              {metric.value}
            </strong>
            <small className={styles.page__metrics__metricCard__metricNote}>{metric.note}</small>
          </article>
        ))}
      </section>

      <h2 className={styles.page__sectionTitle}>Результаты по сценариям</h2>
      <div className={styles.page__tableWrapper}>
        <table className={styles.page__table}>
          <thead>
            <tr className={styles.page__table__header}>
              <th>СЦЕНАРИИ</th>
              <th>ВЕРСИЯ</th>
              <th>ЗАПУСКИ</th>
              <th>ЗАВЕРШИЛИ</th>
              <th>COMPLETION RATE</th>
              <th>GOAL RATE</th>
            </tr>
          </thead>
          <tbody>
            {demoRows.map((row) => (
              <tr
                key={row.id}
                className={styles.page__table__clickableRow}
                onClick={() => navigate(`/admin/analytics/${row.id}`)}
                >
                <td>{row.name}</td>
                <td>{row.version}</td>
                <td>{row.started.toLocaleString("en-US")}</td>
                <td>{row.completed.toLocaleString("en-US")}</td>
                <td className={styles.page__metrics__success}>{row.completionRate.toFixed(1)}%</td>
                <td>{row.goalRate.toFixed(1)}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  );
}