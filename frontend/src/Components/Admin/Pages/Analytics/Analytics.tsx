import styles from "./Analytics.module.scss";

const demoRows = [
  {
    id: "demo-tour-1",
    name: "Второй сценарий",
    version: "v2.1",
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
  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <h1>Дашборд конверсий онбординга</h1>
      </header>

      <section className={styles.metrics} aria-label="Показатели онбординга">
        {demoMetrics.map((metric) => (
          <article className={styles.metricCard} key={metric.label}>
            <span className={styles.metricLabel}>{metric.label}</span>
            <strong className={metric.kind === "success" ? styles.success : metric.kind === "brand" ? styles.brand : undefined}>
              {metric.value}
            </strong>
            <small className={styles.metricNote}>{metric.note}</small>
          </article>
        ))}
      </section>

      <h2 className={styles.sectionTitle}>Результаты по сценариям</h2>
      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <thead>
            <tr className={styles.table__header}>
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
              <tr key={row.id}>
                <td className={styles.tourName}>{row.name}</td>
                <td>{row.version}</td>
                <td>{row.started.toLocaleString("en-US")}</td>
                <td>{row.completed.toLocaleString("en-US")}</td>
                <td className={styles.success}>{row.completionRate.toFixed(1)}%</td>
                <td>{row.goalRate.toFixed(1)}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  );
}