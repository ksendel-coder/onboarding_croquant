import { useEffect, useMemo, useRef, useState } from "react";
import type { ResolveResult, WidgetEventType } from "../../Types";
import { Hint } from "./Hint";
import { makeEvent, resolveTour, sendEvent } from "./client";
import styles from "./OnboardingProvider.module.scss";

interface Props {
  appKey: string;
  apiUrl: string;
}

type RetryAction = "start" | "advance" | "dismiss" | null;

export function OnboardingProvider({ appKey, apiUrl }: Props) {
  const [result, setResult] = useState<ResolveResult | null>(null);
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [retryAction, setRetryAction] = useState<RetryAction>(null);
  const [retryCount, setRetryCount] = useState(0);

  const loadStarted = useRef(false);
  const startPromise = useRef<{
    versionId: string;
    promise: Promise<void>;
  } | null>(null);

  useEffect(() => {
    if (!appKey || loadStarted.current) return;

    loadStarted.current = true;

    void resolveTour(apiUrl, appKey)
      .then(setResult)
      .catch(() => {
        setError("Не удалось загрузить сценарий.");
      });
  }, [apiUrl, appKey]);

  const hints = useMemo(() => [...(result?.tour.hints ?? [])].sort((a, b) => a.step - b.step), [result]);
  const currentIndex = result?.current_hint_id ? Math.max(0, hints.findIndex((h) => h.id === result.current_hint_id)) : 0;
  const hint = hints[currentIndex];

  async function post(type: WidgetEventType, hintId?: string) {
    if (!result) throw new Error("No active tour");
    await sendEvent(apiUrl, appKey, makeEvent(result, type, hintId));
  }

  useEffect(() => {
    if (!result || !hint || ready) return;

    const versionId = result.tour_version_id;

    if (startPromise.current?.versionId !== versionId) {
      startPromise.current = {
        versionId,
        promise: (async () => {
          await sendEvent(apiUrl, appKey, makeEvent(result, "tour_started"));
          await sendEvent(apiUrl, appKey, makeEvent(result, "hint_shown", hint.id));
        })(),
      };
    }

    const pending = startPromise.current;
    let active = true;

    void pending.promise
      .then(() => {
        if (!active) return;

        setReady(true);
        setError(null);
        setRetryAction(null);
      })
      .catch(() => {
        if (startPromise.current === pending) startPromise.current = null;
        if (!active) return;

        setError("Не удалось запустить сценарий.");
        setRetryAction("start");
      });
    return () => {
      active = false;
    };
  }, [result, hint, apiUrl, appKey, retryCount, ready]);

  async function advance() {
    if (!result || !hint || busy || !ready) return;

    setBusy(true);

    try {
      await post("hint_completed", hint.id);

      if (currentIndex === hints.length - 1) {
        await post("tour_completed");
        setResult(null);
      } else {
        const next = hints[currentIndex + 1];
        await post("hint_shown", next.id);
        setResult((prev) => prev ? { ...prev, current_hint_id: next.id } : null);
      }
      setError(null);
      setRetryAction(null);

    } catch {
      setError("Не удалось перейти к следующему шагу.");
      setRetryAction("advance");

    } finally {
      setBusy(false);
    }
  }

  async function dismiss() {
    if (!result || busy || !ready) return;

    setBusy(true);

    try {
      await post("tour_dismissed");

      setResult(null);
      setError(null);
      setRetryAction(null);

    } catch {
      setError("Не удалось закрыть сценарий.");
      setRetryAction("dismiss");

    } finally {
      setBusy(false);
    }
  }

  function retry() {
    setError(null);

    switch (retryAction) {
      case "start":
        setRetryCount((count) => count + 1);
        break;
      case "advance":
        void advance();
        break;
      case "dismiss":
        void dismiss();
        break;
    }
  }

  if (!result && error) {
    return (
      <div className={styles.error} role="alert">
        <p>{error}</p>
        <button type="button" onClick={() => window.location.reload()}>Повторить</button>
      </div>
    );
  }

  if (!appKey || !result || !hint) return null;

  return (
    <>
      {error && (
        <div className={styles.error} role="alert">
          <p>{error}</p>
          <button type="button" disabled={busy} onClick={retry}>Повторить</button>
        </div>
      )}

      <Hint
        hint={hint}
        index={currentIndex}
        total={hints.length}
        onNext={() => void advance()}
        onClose={() => void dismiss()}
        busy={busy || !ready || Boolean(error)}
      />
    </>
  );
}
