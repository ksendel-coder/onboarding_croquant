import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import axios from "axios";
import { useNavigate, useParams } from "react-router-dom";
import { onboardingAPI } from "../../../../Api/onboarding";
import { Hint } from "../../../Onboarding/Hint";
import { WidgetTestHome } from "../../../WidgetTest/WidgetTestHome";
import { WidgetTestProfile } from "../../../WidgetTest/WidgetTestProfile";
import { widgetTestPreviewPath } from "../../../WidgetTest/previewPath";
import siteStyles from "../../../WidgetTest/WidgetTest.module.scss";
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
  const [stepIndex, setStepIndex] = useState(0);
  const [isFinished, setIsFinished] = useState(false);
  const [startedTourId, setStartedTourId] = useState<string>();

  const query = useQuery({
    queryKey: ["tour-preview", tourId],
    queryFn: () => onboardingAPI.getTourPreview(tourId!),
    enabled: Boolean(tourId),
    retry: false,
  });

  const appThemeQuery = useQuery({
    queryKey: ["tour-app-theme", tourId],
    enabled: Boolean(tourId) && query.isSuccess,
    queryFn: async () => {
      const apps = await onboardingAPI.getApps();
      for (const app of apps) {
        const tours = await onboardingAPI.getTours(app.id);
        if (tours?.some((tour) => tour.id === tourId)) return app.theme ?? null;
      }
      return null;
    },
  });

  const hints = useMemo(() => [...(query.data?.tour.hints ?? [])].sort((a, b) => a.step - b.step),[query.data?.tour.hints]);
  const currentIndex = Math.min(stepIndex, Math.max(0, hints.length - 1));
  const currentHint = hints[currentIndex];
  const trigger = query.data?.trigger_type;
  const config = query.data?.trigger_config;
  const scenarioPath = query.data?.tour.target_path;
  const stepPath = hints.slice(0, currentIndex + 1).findLast((hint) => hint.page_path)?.page_path;
  const pagePath = startedTourId === tourId ? stepPath ?? scenarioPath : scenarioPath;
  const testPath = widgetTestPreviewPath(pagePath);
  const ready = query.data?.preview === true && hints.length > 0;
  const started = startedTourId !== undefined && startedTourId === tourId;

  useEffect(() => {
    if (!ready || started || trigger === "manual") return;

    const listeners = new AbortController();
    const options = { passive: true, signal: listeners.signal };
    let timer: number | undefined;
    const start = () => setStartedTourId(tourId);

    switch (trigger) {
      case "delay":
        timer = window.setTimeout(start, config?.delay_ms ?? 1000);
        break;
      case "scroll_depth": {
        const check = () => {
          const page = document.scrollingElement;
          const range = page ? page.scrollHeight - page.clientHeight : 0;
          const depth = page && range > 0 ? (page.scrollTop / range) * 100 : 0;
          if (depth >= (config?.scroll_depth ?? 50)) start();
        };
        window.addEventListener("scroll", check, options);
        window.addEventListener("resize", check, options);
        check();
        break;
      }
      case "inactivity": {
        const reset = () => {
          window.clearTimeout(timer);
          if (!document.hidden) {
            timer = window.setTimeout(start, (config?.inactivity_secs ?? 30) * 1000);
          }
        };
        for (const event of ["pointermove", "pointerdown", "keydown", "scroll"]) {
          window.addEventListener(event, reset, options);
        }
        document.addEventListener("visibilitychange", reset, options);
        reset();
        break;
      }
      case "exit_intent":
        document.addEventListener("mouseout", (event) => {
          if (!event.relatedTarget && event.clientY <= 0) start();
        }, options);
        break;
      default:
        timer = window.setTimeout(start, 0);
    }

    return () => {
      window.clearTimeout(timer);
      listeners.abort();
    };
  }, [ready, started, trigger, config, tourId]);

  const close = () => { navigate("/admin/scenarios"); };

  const handleNext = () => {
    if (currentIndex >= hints.length - 1) {
        setIsFinished(true);
        return;
    }
    setStepIndex(currentIndex + 1);
  };

  const handleSkip = () => { setIsFinished(true); };

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
        <button type="button" onClick={close} data-preview-control="true">Назад к сценариям</button>
      </main>
    );
  }

  return (
    <div className={styles.page}>
      <button type="button" className={styles.page__back} onClick={close} data-preview-control="true">← Назад к сценариям</button>
      <div className={siteStyles.site} data-preview-content="true" onClickCapture={(event) => {
        if (event.target instanceof Element && event.target.closest("a")) event.preventDefault();
      }}>
        {testPath === "/widget-test" && <WidgetTestHome />}
        {testPath === "/widget-test/profile" && <WidgetTestProfile />}
        {!testPath && <p role="status">Страница {pagePath} недоступна на тестовом сайте.</p>}
      </div>

      {hints.length === 0 && <p>В сценарии пока нет шагов.</p>}
      {ready && trigger === "scroll_depth" && (
        <div className={styles.page__scrollSpace} aria-hidden="true" />
      )}
      {ready && !started && trigger === "manual" && (
        <button type="button" className={styles.page__start} onClick={() => setStartedTourId(tourId)} data-preview-control="true">
          Запустить сценарий
        </button>
      )}

      {ready && started && currentHint && !isFinished && (
        <Hint key={currentHint.id} hint={currentHint} index={currentIndex} total={hints.length} onNext={handleNext}
              onMissingTarget={handleNext}
              onClose={handleSkip} busy={false} preview theme={appThemeQuery.data ?? undefined}
        />
      )}
    </div>
  );
}
