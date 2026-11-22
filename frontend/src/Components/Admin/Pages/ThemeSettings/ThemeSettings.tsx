import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { onboardingAPI } from "../../../../Api/onboarding";
import { getOrCreateCurrentApp, selectCurrentApp } from "../../../../Api/helpers/helpers";
import {
  DEFAULT_APP_THEME,
  HEX_COLOR_PATTERN,
  type App,
  type AppTheme,
} from "../../../../Types";
import { getButtonTextColor } from "../../../../Utils/themeColor";
import { useAuth } from "../../../Auth/AuthContext";
import styles from "./ThemeSettings.module.scss";

const FIELDS: { key: keyof AppTheme; title: string; subtitle: string }[] = [
  { key: "primary_color", title: "Акцентный цвет", subtitle: "Рамка, индикатор, кнопка" },
  { key: "background_color", title: "Цвет фона", subtitle: "Фон подсказки" },
  { key: "text_color", title: "Цвет текста", subtitle: "Заголовок и описание" },
];

function readableError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    if (error.response?.status === 400) return "Проверьте HEX-значения цветов.";
    if (error.response?.status === 401) return "Сессия истекла. Войдите снова.";
    if (error.response?.status === 403) return "Нет доступа к приложению.";
    if (error.response?.status === 404) return "Приложение не найдено.";
  }
  return "Не удалось сохранить цвета. Попробуйте ещё раз.";
}

export function ThemeSettings({ appId, disabled }: { appId?: string; disabled: boolean }) {
  const { user } = useAuth();
  const appsKey = ["apps", user?.id] as const;
  const appsQuery = useQuery({
    queryKey: appsKey,
    queryFn: onboardingAPI.getApps,
  });
  const apps = appsQuery.data ?? [];
  const currentApp = appId ? apps.find((app) => app.id === appId) : selectCurrentApp(apps);

  return (
    <details className={styles.settings}>
      <summary>Цвета подсказок</summary>
      {appsQuery.isPending ? (
        <p role="status">Загрузка палитры…</p>
      ) : appsQuery.isError && !appsQuery.data ? (
        <div>
          <p role="alert">Не удалось загрузить палитру.</p>
          <button type="button" onClick={() => void appsQuery.refetch()}>Повторить</button>
        </div>
      ) : appId && !currentApp ? (
        <p role="alert">Приложение сценария недоступно.</p>
      ) : (
        <ThemeEditor app={currentApp ?? null} appsKey={appsKey} disabled={disabled} />
      )}
    </details>
  );
}

function ThemeEditor({ app, appsKey, disabled }: {
  app: App | null;
  appsKey: readonly ["apps", string | undefined];
  disabled: boolean;
}) {
  const queryClient = useQueryClient();
  const [edits, setEdits] = useState<AppTheme | null>(null);
  const persisted = app?.theme ?? DEFAULT_APP_THEME;
  const draft = edits ?? persisted;

  const isValid = FIELDS.every(({ key }) => HEX_COLOR_PATTERN.test(draft[key]));
  const changed = FIELDS.some(({ key }) =>
    draft[key].toUpperCase() !== persisted[key].toUpperCase(),
  );
  const previewColor = (key: keyof AppTheme) => HEX_COLOR_PATTERN.test(draft[key]) ? draft[key] : persisted[key];

  const saveMutation = useMutation({
    mutationKey: ["save-app-theme"],
    mutationFn: async () => {
      if (!isValid) throw new Error("Некорректные данные формы");
      const target = app ?? await getOrCreateCurrentApp();
      return onboardingAPI.updateAppTheme(target.id, {
        primary_color: draft.primary_color.toUpperCase(),
        background_color: draft.background_color.toUpperCase(),
        text_color: draft.text_color.toUpperCase(),
      });
    },
    onSuccess: async (updated) => {
      await queryClient.cancelQueries({ queryKey: appsKey, exact: true });
      queryClient.setQueryData<App[]>(
        appsKey,
        (old = []) => old.some((app) => app.id === updated.id)
          ? old.map((app) => app.id === updated.id ? updated : app)
          : [...old, updated],
      );
      setEdits(null);
    },
  });
  const busy = disabled || saveMutation.isPending;

  function updateColor(key: keyof AppTheme, value: string) {
    setEdits((old) => ({ ...(old ?? persisted), [key]: value }));
    saveMutation.reset();
  }

  return (
    <>
      <p>Палитра общая для всех сценариев приложения.</p>

      <section className={styles.layout}>
        <div className={styles.panel}>
          <h2>Настройка палитры</h2>
          {FIELDS.map(({ key, title, subtitle }) => (
            <div className={styles.field} key={key}>
              <label htmlFor={`theme-${key}`}>{title}</label>
              <small>{subtitle}</small>
              <div className={styles.controls}>
                <input
                  type="color"
                  aria-label={`Выбрать: ${title}`}
                  value={previewColor(key)}
                  onInput={(event) => updateColor(key, event.currentTarget.value.toUpperCase())}
                  disabled={busy}
                />
                <input
                  id={`theme-${key}`}
                  type="text"
                  value={draft[key]}
                  maxLength={7}
                  autoComplete="off"
                  spellCheck={false}
                  aria-invalid={!HEX_COLOR_PATTERN.test(draft[key])}
                  onChange={(event) => updateColor(key, event.target.value)}
                  disabled={busy}
                />
              </div>
              {!HEX_COLOR_PATTERN.test(draft[key]) && (
                <small className={styles.error}>Нужен формат #RRGGBB, например #696CF1</small>
              )}
            </div>
          ))}

          <div className={styles.actions}>
            <button type="button" disabled={busy}
              onClick={() => { setEdits(DEFAULT_APP_THEME); saveMutation.reset(); }}>
              Сбросить к стандартным
            </button>
            <button type="button" disabled={busy || !isValid || !changed}
              onClick={() => saveMutation.mutate()}>
              {saveMutation.isPending ? "Сохранение..." : "Сохранить цвета"}
            </button>
          </div>
          <p className={styles.note}>Чтобы применить стандартные цвета, нажмите «Сохранить цвета» после сброса.</p>
          {saveMutation.isError && <p role="alert" className={styles.error}>{readableError(saveMutation.error)}</p>}
          {saveMutation.isSuccess && <p role="status" className={styles.success}>Цвета сохранены.</p>}
        </div>

        <div className={styles.panel}>
          <h2>Предпросмотр</h2>
          <div className={styles.previewStage}>
            <article className={styles.hint} style={{
              backgroundColor: previewColor("background_color"),
              color: previewColor("text_color"),
              borderColor: previewColor("primary_color"),
            }}>
              <div className={styles.hintHeader}>
                <strong>Добро пожаловать!</strong>
                <span style={{ color: previewColor("primary_color") }}>1/3</span>
              </div>
              <p>Нажмите «Далее», чтобы познакомиться с интерфейсом.</p>
              <button type="button" tabIndex={-1} aria-disabled="true" style={{
                backgroundColor: previewColor("primary_color"),
                color: getButtonTextColor(previewColor("primary_color")),
              }}>Далее</button>
            </article>
          </div>
        </div>
      </section>
    </>
  );
}
