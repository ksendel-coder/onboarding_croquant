import { useQueries } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { publishedTourAnalyticsQuery } from "../../../../Hooks/publishedTourAnalyticsQuery";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import styles from "./Analytics.module.scss";

export function Analytics() {
  const navigate = useNavigate();
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
    { label: "ВСЕГО ЗАПУСКОВ", value: started.toLocaleString("ru-RU"), note: `Версий в сводке — ${rows.length.toLocaleString("ru-RU")}` },
    { label: "ЗАВЕРШИЛИ ТУР", value: completed.toLocaleString("ru-RU"), note: "Дошли до конца и выполнили хотя бы один шаг" },
    {
      label: "ПРОПУСТИЛИ ТУР",
      value: dismissed.toLocaleString("ru-RU"),
      note: started > 0 ? `${((dismissed * 100) / started).toFixed(1)}% от всех запусков` : "Пока нет запусков",
    },
    {
      label: "COMPLETION RATE",
      value: `${completionRate.toFixed(1)}%`,
      note: started > 0 ? `${completed.toLocaleString("ru-RU")} из ${started.toLocaleString("ru-RU")} запусков` : "Пока нет запусков",
      emphasis: completionRate > 0,
    },
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
                  <strong className={metric.emphasis ? styles.page__metrics__success : undefined}>{metric.value}</strong>
                  <small className={styles.page__metrics__metricCard__metricNote}>{metric.note}</small>
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
                  const detailsPath = `/admin/analytics/${tour.id}`;

                  return (
                    <tr className={styles.page__table__row} key={tour.id} onClick={(event) => {
                      if (event.target instanceof Element && !event.target.closest("a")) navigate(detailsPath);
                    }}>
                      <td><Link to={detailsPath}>{tour.title}</Link></td>
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
