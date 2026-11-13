import { useEffect, useMemo, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import axios from "axios";
import { useNavigate, useParams } from "react-router-dom";
import { onboardingAPI } from "../../../../Api/onboarding";
import { Hint } from "../../../Onboarding/Hint";
import styles from "./ScenarioPreview.module.scss";

function errorText(error: unknown): string {
  if (axios.isAxiosError(error)) {
    switch (error.response?.status) {
      case 401:
        return "Сессия истекла. Войдите снова.";
      case 403:
        return "Нет доступа к сценарию.";
      case 404:
        return "Сценарий не найден.";
      case 409:
        return "У сценария нет черновика для предпросмотра.";
    }
  }
  return "Не удалось загрузить предпросмотр.";
}

export function ScenarioPreview() {
  const { tourId } = useParams<{ tourId: string }>();
  const navigate = useNavigate();
  const pageRef = useRef<HTMLDivElement>(null);
  const [stepIndex, setStepIndex] = useState(0);
  const [isFinished, setIsFinished] = useState(false);
  const [pageHtml, setPageHtml] = useState<string | null>(null);
  const [pageError, setPageError] = useState(false);

  const query = useQuery({
    queryKey: ["tour-preview", tourId],
    queryFn: () => onboardingAPI.getTourPreview(tourId!),
    enabled: Boolean(tourId),
    retry: false,
  });

  const hints = useMemo(() => [...(query.data?.tour.hints ?? [])].sort((a, b) => a.step - b.step),[query.data?.tour.hints]);
  const currentIndex = Math.min(stepIndex, Math.max(0, hints.length - 1));
  const currentHint = hints[currentIndex];

  const close = () => { navigate("/admin/scenarios"); };

  const handleNext = () => {
    if (currentIndex >= hints.length - 1) {
        setIsFinished(true);
        return;
    }
    setStepIndex((prev) => prev + 1);
  };

  const handleSkip = () => { setIsFinished(true); };

  useEffect(() => {
    const controller = new AbortController();

    async function loadPage() {
      setPageHtml(null);
      setPageError(false);

      try {
        const response = await fetch("/picker-preview.html", { signal: controller.signal });
        if (!response.ok) throw new Error("Не удалось загрузить тестовую страницу");

        const doc = new DOMParser().parseFromString(await response.text(), "text/html");
        doc.querySelectorAll("script").forEach((script) => { script.remove(); });

        if (!controller.signal.aborted) setPageHtml(doc.body.innerHTML);
      } catch {
        if (!controller.signal.aborted) setPageError(true);
      }
    }
    void loadPage();
    return () => { controller.abort(); };
  }, []);

  if (query.isPending) {
    return (
      <main className={styles.page}>
        <p>Загружаем предпросмотр...</p>
      </main>
    );
  }

  if (query.isError) {
    return (
      <main className={styles.page}>
        <p role="alert">{errorText(query.error)}</p>
        <button type="button" onClick={() => void query.refetch()}>Повторить запрос</button>
        <button type="button" onClick={close}>Назад к сценариям</button>
      </main>
    );
  }

  if (!query.data.preview) {
    return (
      <main className={styles.page}>
        <p>Backend не подтвердил режим предпросмотра.</p>
        <button type="button" onClick={close}>Назад</button>
      </main>
    );
  }

  if (pageError) {
    return (
      <main className={styles.page}>
        <p>Не удалось загрузить тестовую страницу.</p>
        <button type="button" onClick={close}>Назад к сценариям</button>
      </main>
    );
  }

  return (
    <main className={styles.page}>
      <button type="button" className={styles.page__back} onClick={close}>← Назад к сценариям</button>
      <div
        ref={pageRef}
        className={styles.page__content}
        dangerouslySetInnerHTML={ pageHtml !== null ? { __html: pageHtml } : undefined}
      />

      {pageHtml === null && <p>Загружаем тестовую страницу...</p>}
      {pageHtml !== null && hints.length === 0 && (<p>В сценарии пока нет шагов.</p>)}

      {pageHtml !== null && currentHint && !isFinished && (
        <>
          <Hint key={currentHint.id} hint={currentHint} index={currentIndex} total={hints.length} onNext={handleNext}
                onClose={handleSkip} busy={false} preview
          />
        </>
      )}
    </main>
  );
}
