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

  if (tours.isError) {
    return <main className={styles.page}><p role="alert">Не удалось загрузить аналитику.</p></main>;
  }

  const entries = scenarios.map((tour, index) => ({ tour, report: reports[index] }));
  const rows = entries.flatMap(({ report }) => report.data ? [report.data] : []);
  const hasMissingReports = entries.some(({ report }) => report.isError && !report.data);
  const started = rows.reduce((total, row) => total + row.analytics.started, 0);
  const completed = rows.reduce((total, row) => total + row.analytics.completed, 0);
  const dismissed = rows.reduce((total, row) => total + row.analytics.dismissed, 0);
  const completionRate = started > 0 ? (100 * completed) / started : 0;
  const metrics = [
    { label: "ВСЕГО ЗАПУСКОВ", value: started.toLocaleString("ru-RU") },
    { label: "ЗАВЕРШИЛИ ТУР", value: completed.toLocaleString("ru-RU") },
    { label: "ПРОПУСТИЛИ ТУР", value: dismissed.toLocaleString("ru-RU") },
    { label: "COMPLETION RATE", value: `${completionRate.toFixed(1)}%` },
  ];

  return (
    <main className={styles.page}>
      <header className={styles.page__header}>
        <h1>Дашборд конверсий онбординга</h1>
      </header>

      {rows.length === 0 && !hasMissingReports ? (
        <p>Опубликуйте сценарий, чтобы увидеть его аналитику.</p>
      ) : (
        <>
          {hasMissingReports && rows.length > 0 && (
            <p role="status">Сводка рассчитана по сценариям, для которых удалось загрузить данные.</p>
          )}
          {rows.length > 0 && (
            <section className={styles.page__metrics} aria-label="Показатели онбординга">
              {metrics.map((metric) => (
                <article className={styles.page__metrics__metricCard} key={metric.label}>
                  <span className={styles.page__metrics__metricCard__metricLabel}>{metric.label}</span>
                  <strong>{metric.value}</strong>
                </article>
              ))}
            </section>
          )}

          <h2 className={styles.page__sectionTitle}>Результаты по текущим опубликованным версиям</h2>
          <div className={styles.page__tableWrapper}>
            <table className={styles.page__table}>
              <thead>
                <tr className={styles.page__table__header}>
                  <th>СЦЕНАРИИ</th>
                  <th>ВЕРСИЯ</th>
                  <th>ЗАПУСКИ</th>
                  <th>ЗАВЕРШИЛИ</th>
                  <th>ПРОПУСТИЛИ</th>
                  <th>COMPLETION RATE</th>
                </tr>
              </thead>
              <tbody>
                {entries.map(({ tour, report }) => {
                  if (!report.data && !report.isError) return null;

                  return (
                    <tr key={tour.id}>
                      <td><Link to={`/admin/analytics/${tour.id}`}>{tour.title}</Link></td>
                      {report.data ? (
                        <>
                          <td>v{report.data.version.version}</td>
                          <td>{report.data.analytics.started.toLocaleString("ru-RU")}</td>
                          <td>{report.data.analytics.completed.toLocaleString("ru-RU")}</td>
                          <td>{report.data.analytics.dismissed.toLocaleString("ru-RU")}</td>
                          <td className={styles.page__metrics__success}>{report.data.analytics.completion_rate.toFixed(1)}%</td>
                        </>
                      ) : (
                        <td colSpan={5}>Не удалось загрузить статистику.</td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </main>
  );
}
