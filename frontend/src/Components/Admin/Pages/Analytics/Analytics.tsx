import { useQueries } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { publishedTourAnalyticsQuery } from "../../../../Hooks/publishedTourAnalyticsQuery";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import styles from "./Analytics.module.scss";

export function Analytics() {
  const tours = useToursQuery();
  const scenarios = tours.data ?? [];
  const reports = useQueries({
    queries: scenarios.map((tour) => publishedTourAnalyticsQuery(tour.id)),
  });

  if (tours.isPending || reports.some((report) => report.isPending)) {
    return <main className={styles.page}><p role="status">Загрузка аналитики…</p></main>;
  }

  if (tours.isError || reports.some((report) => report.isError)) {
    return <main className={styles.page}><p role="alert">Не удалось загрузить аналитику.</p></main>;
  }

  const rows = scenarios.flatMap((tour, index) => {
    const report = reports[index].data;
    return report ? [{ tour, ...report }] : [];
  });
  const started = rows.reduce((total, row) => total + row.analytics.started, 0);
  const completed = rows.reduce((total, row) => total + row.analytics.completed, 0);
  const completionRate = started > 0 ? (100 * completed) / started : 0;
  const metrics = [
    { label: "ВСЕГО ЗАПУСКОВ", value: started.toLocaleString("ru-RU") },
    { label: "ЗАВЕРШИЛИ ТУР", value: completed.toLocaleString("ru-RU") },
    { label: "COMPLETION RATE", value: `${completionRate.toFixed(1)}%` },
  ];

  return (
    <main className={styles.page}>
      <header className={styles.page__header}>
        <h1>Дашборд конверсий онбординга</h1>
      </header>

      {rows.length === 0 ? (
        <p>Опубликуйте сценарий, чтобы увидеть его аналитику.</p>
      ) : (
        <>
          <section className={styles.page__metrics} aria-label="Показатели онбординга">
            {metrics.map((metric) => (
              <article className={styles.page__metrics__metricCard} key={metric.label}>
                <span className={styles.page__metrics__metricCard__metricLabel}>{metric.label}</span>
                <strong>{metric.value}</strong>
              </article>
            ))}
          </section>

          <h2 className={styles.page__sectionTitle}>Результаты по опубликованным версиям</h2>
          <div className={styles.page__tableWrapper}>
            <table className={styles.page__table}>
              <thead>
                <tr className={styles.page__table__header}>
                  <th>СЦЕНАРИИ</th>
                  <th>ВЕРСИЯ</th>
                  <th>ЗАПУСКИ</th>
                  <th>ЗАВЕРШИЛИ</th>
                  <th>COMPLETION RATE</th>
                </tr>
              </thead>
              <tbody>
                {rows.map(({ tour, version, analytics }) => (
                  <tr key={tour.id}>
                    <td><Link to={`/admin/analytics/${tour.id}`}>{tour.title}</Link></td>
                    <td>v{version.version}</td>
                    <td>{analytics.started.toLocaleString("ru-RU")}</td>
                    <td>{analytics.completed.toLocaleString("ru-RU")}</td>
                    <td className={styles.page__metrics__success}>{analytics.completion_rate.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </main>
  );
}
