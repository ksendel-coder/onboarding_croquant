import { useQueries } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { publishedTourAnalyticsQuery } from "../../../../Hooks/publishedTourAnalyticsQuery";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import type { AnalyticsPeriod, TourAnalytics } from "../../../../Types";
import styles from "./Analytics.module.scss";

const trendDescription = "Последние 7 дней по сравнению с предыдущими 7 днями";
const conversionDescription = "Сравнение конверсии за последние 7 дней со средней конверсией за предыдущие 7 дней";

function trend(current: number, previous: number, available: boolean, fallback: { note: string; description: string }) {
  if (!available || previous === 0) {
    return {
      ...fallback,
      note: available && current > 0 ? `↑ ${fallback.note}` : fallback.note,
    };
  }
  const change = ((current - previous) / previous) * 100;
  if (change === 0) return { note: "0,0%", description: trendDescription };
  const percent = Math.abs(change).toLocaleString("ru-RU", { minimumFractionDigits: 1, maximumFractionDigits: 1 });
  return { note: `${change > 0 ? "↑ +" : "↓ −"}${percent}%`, description: trendDescription };
}

function share(count: number, total: number, label: string) {
  const note = (total ? count / total : 0).toLocaleString("ru-RU", { style: "percent", maximumFractionDigits: 1 });
  return { note, description: total ? label : "Пока нет запусков" };
}

function sumPeriod(rows: { analytics: TourAnalytics }[], period: "recent" | "previous") {
  return rows.reduce<AnalyticsPeriod>((total, { analytics }) => {
    const counts = analytics[period];
    return counts ? {
      started: total.started + counts.started,
      completed: total.completed + counts.completed,
      dismissed: total.dismissed + counts.dismissed,
    } : total;
  }, { started: 0, completed: 0, dismissed: 0 });
}

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
  const recent = sumPeriod(rows, "recent");
  const previous = sumPeriod(rows, "previous");
  const hasPeriods = rows.every(({ analytics }) => analytics.recent && analytics.previous);
  const canCompare = !hasMissingReports && hasPeriods;
  const compareWeeks = canCompare && recent.started > 0 && previous.started > 0;
  const activeReports = rows.filter(({ analytics }) => analytics.started > 0);
  const averageRate = activeReports.reduce((total, { analytics }) => total + analytics.completion_rate, 0) / Math.max(1, activeReports.length);
  const conversionDifference = compareWeeks
    ? (recent.completed / recent.started - previous.completed / previous.started) * 100
    : completionRate - averageRate;
  const metrics: { label: string; value: string; note: string; description?: string; emphasis?: boolean }[] = [
    {
      label: "ВСЕГО ЗАПУСКОВ",
      value: started.toLocaleString("ru-RU"),
      ...trend(recent.started, previous.started, canCompare, hasPeriods ? share(recent.started, started, "Доля запусков за последние 7 дней") : share(started, started, "Доля запусков в сводке")),
    },
    { label: "ЗАВЕРШИЛИ ТУР", value: completed.toLocaleString("ru-RU"), ...trend(recent.completed, previous.completed, canCompare, share(completed, started, "Доля завершений от всех запусков")) },
    {
      label: "ПРОПУСТИЛИ ТУР",
      value: dismissed.toLocaleString("ru-RU"),
      ...trend(recent.dismissed, previous.dismissed, canCompare, share(dismissed, started, "Доля пропусков от всех запусков")),
    },
    {
      label: "COMPLETION RATE",
      value: `${completionRate.toFixed(1)}%`,
      note: !started ? "Пока нет запусков" : Math.abs(conversionDifference) < 0.05 ? "На уровне среднего"
        : conversionDifference > 0 ? "Выше среднего" : "Ниже среднего",
      description: compareWeeks ? conversionDescription : "Сравнение со средним показателем по сценариям с запусками в сводке",
      emphasis: true,
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
                  <small className={styles.page__metrics__metricCard__metricNote} title={metric.description} data-percent={!metric.emphasis}>{metric.note}</small>
                </article>
              ))}
            </section>
          )}

          <h2 className={styles.page__sectionTitle}>Результаты по сценариям</h2>
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
