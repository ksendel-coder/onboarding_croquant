import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, FileText } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { publishedTourAnalyticsQuery } from "../../../../Hooks/publishedTourAnalyticsQuery";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import styles from "./AnalyticsDetail.module.scss";

export function AnalyticsDetail() {
  const { tourId } = useParams<{ tourId: string }>();
  const tours = useToursQuery();
  const tour = tours.data?.find((item) => item.id === tourId);
  const report = useQuery({
    ...publishedTourAnalyticsQuery(tourId ?? ""),
    enabled: Boolean(tour),
  });

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
      ) : !report.data ? (
        <p>У сценария пока нет опубликованной версии.</p>
      ) : (
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

          <h2>Аналитика по шагам</h2>
          <p className={styles.panel__description}>
            Запусков: {report.data.analytics.started} · Завершений: {report.data.analytics.completed} ·
            Закрытий: {report.data.analytics.dismissed}
          </p>

          {report.data.analytics.steps.length === 0 ? (
            <p>В опубликованной версии нет шагов.</p>
          ) : (
            <div className={styles.panel__steps}>
              {report.data.analytics.steps.map((step) => (
                <article className={styles.panel__steps__stepRow} key={step.hint_id}>
                  <div className={styles.panel__steps__stepInfo}>
                    <span className={styles.panel__steps__stepNumber}>Шаг {step.step}</span>
                    {step.title && <h3>{step.title}</h3>}
                  </div>
                  <div className={styles.panel__steps__stat}><span>Показано</span><strong>{step.shown}</strong></div>
                  <div className={styles.panel__steps__stat}><span>Завершено</span><strong>{step.completed}</strong></div>
                  <div className={styles.panel__steps__stat}><span>Автопропуски</span><strong>{step.skipped ?? "—"}</strong></div>
                  <div className={styles.panel__steps__stat}><span>Показано без завершения</span><strong>{step.drop_off}</strong></div>
                </article>
              ))}
            </div>
          )}
        </section>
      )}
    </main>
  );
}
