import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ChevronRight, Copy, FileText, ImageOff } from "lucide-react";
import styles from "./AnalyticsDetail.module.scss";

const demoTour = {
  id: "demo-tour-1",
  name: "Публикация",
  version: 7,
  period: "15.07.2026 — 14.08.2026",
  versionId: "2f15b19d-f52d-450a-b06a-b9f8ff1ec11e",
  steps: [
    {
      id: "step-1",
      number: 1,
      title: "Первое объявление",
      description: "Создание и публикация первого объявления",
      shown: 1,
      completed: 1,
      skipped: 0,
      missing: 0,
    },
    {
      id: "step-2",
      number: 2,
      title: "Цена",
      description: "Установка цены товара",
      shown: 1,
      completed: 1,
      skipped: 0,
      missing: 0,
    },
    {
      id: "step-3",
      number: 3,
      title: "Фото товара",
      description: "Загрузка фотографий товара",
      shown: 0,
      completed: 0,
      skipped: 1,
      missing: 1,
    },
  ],
  issues: [
    {
      id: "problem-1",
      step: 3,
      name: "Фото товара",
      description: "Не найден селектор элемента на странице",
      count: 1,
      rate: 100,
      lastSeen: "14.08.2026, 12:34",
    },
  ],
};

function percent(value: number, total: number) {
  return total > 0 ? Math.round((value / total) * 100) : 0;
}

export function AnalyticsDetail() {
  const { tourId } = useParams<{ tourId: string }>();
  const [copied, setCopied] = useState(false);

  if (tourId !== demoTour.id) {
    return (
      <main className={styles.page}>
        <p>Сценарий не найден.</p>
        <Link to="/admin/analytics">Вернуться к аналитике</Link>
      </main>
    );
  }

  async function copyVersionId() {
    try {
      await navigator.clipboard.writeText(demoTour.versionId);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }

  return (
    <main className={styles.page}>
      <section className={styles.panel}>
        <header className={styles.panel__top}>
          <div className={styles.panel__tourHeading}>
            <span className={styles.panel__tourHeading__documentIcon} aria-hidden="true"><FileText size={28} /></span>
            <div>
              <h1>{demoTour.name}</h1>
              <p>Версия {demoTour.version} <span className={styles.page__dot}>·</span> {demoTour.period}</p>
            </div>
          </div>
          <div className={styles.page__versionBox}>
            <div>
              <span>Version ID</span>
              <strong>{demoTour.versionId}</strong>
            </div>
            <button type="button" className={styles.page__copyButton} onClick={() => void copyVersionId()}
              title="Скопировать Version ID" aria-label="Скопировать Version ID">
              <Copy size={20} />
            </button>
            {copied && <span className={styles.panel__copied}>Скопировано</span>}
          </div>
        </header>

        <h2>Аналитика</h2>
        <p className={styles.panel__description}>Статистика по ключевым шагам публикации. Отслеживайте, на каком этапе 
            пользователи чаще всего останавливаются.
        </p>

        <div className={styles.panel__steps}>
          {demoTour.steps.map((step) => {
            const completedRate = percent(step.completed, step.shown);
            const skippedRate = percent(step.skipped, Math.max(step.shown, step.skipped));
            const missingRate = percent(step.missing, Math.max(step.shown, step.missing));
            return (
              <article className={styles.panel__steps__stepRow} key={step.id}>
                <div className={styles.panel__steps__stepInfo}>
                  <span className={styles.panel__steps__stepNumber}>{step.number}</span>
                  <div>
                    <h3>{step.title}</h3>
                    <p>{step.description}</p>
                  </div>
                </div>
                <div className={styles.panel__steps__stat}><span>Показано</span><strong>{step.shown}</strong></div>
                <div className={styles.panel__steps__stat}>
                  <span>Завершено</span>
                  <strong>{step.completed} <small className={step.completed > 0 ? styles.panel__steps__stat__good 
                    : styles.panel__steps__stat__neutral}>{completedRate}%</small>
                  </strong>
                </div>
                <div className={styles.panel__steps__stat}>
                  <span>Пропущено</span>
                  <strong>{step.skipped} <small className={step.skipped > 0 ? styles.panel__steps__stat__bad 
                    : styles.panel__steps__stat__neutral}>{skippedRate}%</small>
                  </strong>
                </div>
                <div className={styles.panel__steps__stat}>
                  <span>Проблемные селекторы</span>
                  <strong>{step.missing} <small className={step.missing > 0 ? styles.panel__steps__stat__bad 
                    : styles.panel__steps__stat__neutral}>{missingRate}%</small>
                  </strong>
                </div>
              </article>
            );
          })}
        </div>
      </section>

      <section className={styles.panel}>
        <h2>Проблемные селекторы <span className={styles.panel__issueCount}>{demoTour.issues.length}</span></h2>
        <p className={styles.panel__description}>Элементы, с которыми пользователи чаще всего сталкиваются с проблемами.</p>
        <div className={styles.panel__issues}>
          {demoTour.issues.map((issue) => (
            <article className={styles.panel__issues__issueRow} key={issue.id}>
              <span className={styles.panel__issues__issueIcon}><ImageOff size={27} /></span>
              <div className={styles.panel__issues__issueInfo}>
                <h3>Шаг {issue.step} — {issue.name}</h3>
                <p>{issue.description}</p>
              </div>
              <div className={styles.panel__issueStat}>
                <span>Количество ошибок</span>
                <strong>{issue.count} <small className={styles.panel__issueStat__bad}>{issue.rate}%</small></strong>
              </div>
              <div className={styles.panel__issueStat}>
                <span>Последнее вхождение</span>
                <strong>{issue.lastSeen}</strong>
              </div>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}