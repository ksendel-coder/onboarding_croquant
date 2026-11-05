import { useEffect, useMemo, useRef, useState } from "react";
import type { ResolveResult, WidgetEventType } from "../../Types";
import { Hint } from "./Hint";
import { makeEvent, resolveTour, sendEvent } from "./client";
import styles from "./OnboardingProvider.module.scss";

interface Props { appKey: string; apiUrl: string }

export function OnboardingProvider({ appKey, apiUrl }: Props) {
  const [result, setResult] = useState<ResolveResult | null>(null);
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const initialized = useRef<string | null>(null);
  const loadStarted = useRef(false);

  useEffect(() => {
    if (!appKey || loadStarted.current) return;

    loadStarted.current = true;
    void resolveTour(apiUrl, appKey)
      .then(setResult)
      .catch((err: unknown) => setError(String(err)));
  }, [apiUrl, appKey]);

  const hints = useMemo(() => [...(result?.tour.hints ?? [])].sort((a, b) => a.step - b.step), [result]);
  const currentIndex = result?.current_hint_id ? Math.max(0, hints.findIndex((h) => h.id === result.current_hint_id)) : 0;
  const hint = hints[currentIndex];

  async function post(type: WidgetEventType, hintId?: string) {
    if (!result) throw new Error("No active tour");
    await sendEvent(apiUrl, appKey, makeEvent(result, type, hintId));
  }

  useEffect(() => {
    if (!result || !hint || initialized.current === result.tour_version_id) return;
    initialized.current = result.tour_version_id;

    let cancelled = false;

    const start = async () => {
      try {
        await sendEvent(apiUrl, appKey, makeEvent(result, "tour_started"));
        await sendEvent(apiUrl, appKey, makeEvent(result, "hint_shown", hint.id));

        if (!cancelled) {
          setReady(true);
          setError(null);
        }
      } catch {
        if (!cancelled) {
          initialized.current = null;
          setError("Не удалось запустить сценарий. Попробуйте ещё раз.");
        }
      }
    };
    void start();

    return () => {
      cancelled = true;

      if (initialized.current === result.tour_version_id && !ready) initialized.current = null;
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
        setResult((prev) => prev && ({ ...prev, current_hint_id: next.id }));
      }
    } catch (err) { setError(String(err)); }
    finally { setBusy(false); }
  }

  async function dismiss() {
    if (busy || !ready) return;
    
    setBusy(true);
    try { await post("tour_dismissed"); setResult(null); }
    catch (err) { setError(String(err)); }
    finally { setBusy(false); }
  }

  if (!result && error) {
    return (
      <div role="alert">
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
          <button type="button" disabled={busy} onClick={() => {setError(null); setRetryCount((count) => count + 1);}}>
            Повторить
          </button>
        </div>
      )}
      <Hint hint={hint} index={currentIndex} total={hints.length}
        onNext={() => void advance()} onClose={() => void dismiss()}
        busy={busy || !ready} />
    </>
  );
}