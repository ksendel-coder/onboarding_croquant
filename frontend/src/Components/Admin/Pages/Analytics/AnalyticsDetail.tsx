import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, CircleAlert, FileText } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { publishedTourAnalyticsQuery } from "../../../../Hooks/publishedTourAnalyticsQuery";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import styles from "./AnalyticsDetail.module.scss";

function percent(count: number, total: number) {
  return total > 0 ? `${Math.round((count / total) * 100)}%` : "—";
}

export function AnalyticsDetail() {
  const { tourId } = useParams<{ tourId: string }>();
  const tours = useToursQuery();
  const tour = tours.data?.find((item) => item.id === tourId);
  const report = useQuery({
    ...publishedTourAnalyticsQuery(tourId ?? ""),
    enabled: Boolean(tour),
  });
  const analytics = report.data?.analytics;
  const skippedSteps = analytics?.steps.filter((step) => step.skipped > 0) ?? [];

  return (
    <main className={styles.page}>
      <Link className={styles.page__back} to="/admin/analytics">
        <ArrowLeft size={18} aria-hidden="true" />
        Вернуться к аналитике
      </Link>

      {tours.isPending || (tour && report.isPending) ? (
        <p role="status">Загрузка аналитики…</p>
      ) : tours.isError || report.isError ? (
        <p role="alert">Не удалось загрузить аналитику.</p>
      ) : !tour ? (
        <p>Сценарий не найден.</p>
      ) : !report.data || !analytics ? (
        <p>У сценария пока нет опубликованной версии.</p>
      ) : (
        <>
          <section className={styles.panel}>
            <header className={styles.panel__top}>
              <div className={styles.panel__tourHeading}>
                <span className={styles.panel__tourHeading__documentIcon} aria-hidden="true"><FileText size={28} /></span>
                <div>
                  <h1>{tour.title}</h1>
                  <p>Версия {report.data.version.version}</p>
                </div>
              </div>
            </header>

            <h2>Аналитика</h2>
            <p className={styles.panel__description}>
              Статистика по шагам сценария
            </p>

            {analytics.steps.length === 0 ? (
              <p>В опубликованной версии нет шагов.</p>
            ) : (
              <div className={styles.panel__steps}>
                {analytics.steps.map((step) => (
                  <article className={styles.panel__steps__stepRow} key={step.hint_id}>
                    <div className={styles.panel__steps__stepInfo}>
                      <span className={styles.panel__steps__stepNumber}>{step.step}</span>
                      <h3>{step.title || `Шаг ${step.step}`}</h3>
                    </div>
                    <div className={styles.panel__steps__stat}><span>Показано</span><strong>{step.shown}</strong></div>
                    <div className={styles.panel__steps__stat}>
                      <span>Завершено</span>
                      <strong>{step.completed} <small className={styles.panel__steps__stat__good}>{percent(step.completed, step.shown)}</small></strong>
                    </div>
                    <div className={styles.panel__steps__stat}><span>Автопропуски</span><strong>{step.skipped ?? "—"}</strong></div>
                    <div className={styles.panel__steps__stat}><span>Показано без завершения</span><strong>{step.drop_off}</strong></div>
                  </article>
                ))}
              </div>
            )}
          </section>
          <section className={styles.panel}>
            <h2>Проблемные селекторы <span className={styles.panel__issueCount}>{skippedSteps.length}</span></h2>
            {skippedSteps.length === 0 ? <p>Автопропусков пока нет.</p> : (
              <div className={styles.panel__issues}>
                {skippedSteps.map((step) => (
                  <article className={styles.panel__issues__issueRow} key={step.hint_id}>
                    <span className={styles.panel__issues__issueIcon} aria-hidden="true"><CircleAlert size={27} /></span>
                    <div className={styles.panel__issues__issueInfo}>
                      <h3>Шаг {step.step}{step.title && ` — ${step.title}`}</h3>
                      <p>Целевой элемент был недоступен</p>
                    </div>
                    <div className={styles.panel__issueStat}>
                      <span>Автопропуски</span>
                      <strong>{step.skipped}</strong>
                    </div>
                  </article>
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </main>
  );
}
