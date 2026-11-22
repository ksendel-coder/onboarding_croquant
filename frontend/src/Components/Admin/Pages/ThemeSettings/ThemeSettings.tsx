import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { onboardingAPI } from "../../../../Api/onboarding";
import {
  DEFAULT_APP_THEME,
  HEX_COLOR_PATTERN,
  type App,
  type AppTheme,
} from "../../../../Types";
import { getButtonTextColor } from "../../../../Utils/themeColor";
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

export function ThemeSettings() {
  const appsQuery = useQuery({
    queryKey: ["apps"],
    queryFn: onboardingAPI.getApps,
  });
  const apps = appsQuery.data ?? [];
  const [appId, setAppId] = useState("");
  const currentApp = apps.find((app) => app.id === appId) ?? apps[0];

  if (appsQuery.isPending) return <main className={styles.page}><p role="status">Загрузка приложений…</p></main>;
  if (appsQuery.isError && !currentApp) return (
    <main className={styles.page}>
      <p role="alert">Не удалось загрузить приложения.</p>
      <button type="button" onClick={() => void appsQuery.refetch()}>Повторить</button>
    </main>
  );
  if (!currentApp) return (
    <main className={styles.page}>
      <h1>Цвета подсказок</h1>
      <p>Сначала создайте приложение в административной панели.</p>
    </main>
  );

  return <ThemeEditor key={currentApp.id} app={currentApp} apps={apps} onSelect={setAppId} />;
}

function ThemeEditor({ app, apps, onSelect }: { app: App; apps: App[]; onSelect: (id: string) => void }) {
  const queryClient = useQueryClient();
  const [edits, setEdits] = useState<AppTheme | null>(null);
  const persisted = app.theme ?? DEFAULT_APP_THEME;
  const draft = edits ?? persisted;

  const isValid = FIELDS.every(({ key }) => HEX_COLOR_PATTERN.test(draft[key]));
  const changed = FIELDS.some(({ key }) =>
    draft[key].toUpperCase() !== persisted[key].toUpperCase(),
  );
  const previewColor = (key: keyof AppTheme) => HEX_COLOR_PATTERN.test(draft[key]) ? draft[key] : persisted[key];

  const saveMutation = useMutation({
    mutationFn: () => {
      if (!isValid) throw new Error("Некорректные данные формы");
      return onboardingAPI.updateAppTheme(app.id, {
        primary_color: draft.primary_color.toUpperCase(),
        background_color: draft.background_color.toUpperCase(),
        text_color: draft.text_color.toUpperCase(),
      });
    },
    onSuccess: async (updated) => {
      await queryClient.cancelQueries({ queryKey: ["apps"], exact: true });
      queryClient.setQueryData<App[]>(
        ["apps"],
        (old) => old?.map((app) => app.id === updated.id ? updated : app),
      );
      setEdits(null);
    },
  });

  function updateColor(key: keyof AppTheme, value: string) {
    setEdits((old) => ({ ...(old ?? persisted), [key]: value }));
    saveMutation.reset();
  }

  return (
    <main className={styles.page}>
      <h1>Цвета подсказок</h1>
      <p>Палитра применяется ко всем опубликованным сценариям выбранного приложения.</p>

      <label className={styles.appLabel} htmlFor="theme-app">Приложение</label>
      <select
        id="theme-app"
        value={app.id}
        disabled={saveMutation.isPending}
        onChange={(event) => onSelect(event.target.value)}
      >
        {apps.map((app) => <option key={app.id} value={app.id}>{app.name}</option>)}
      </select>

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
                  onChange={(event) => updateColor(key, event.target.value.toUpperCase())}
                  disabled={saveMutation.isPending}
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
                  disabled={saveMutation.isPending}
                />
              </div>
              {!HEX_COLOR_PATTERN.test(draft[key]) && (
                <small className={styles.error}>Нужен формат #RRGGBB, например #696CF1</small>
              )}
            </div>
          ))}

          <div className={styles.actions}>
            <button type="button" disabled={saveMutation.isPending}
              onClick={() => { setEdits(DEFAULT_APP_THEME); saveMutation.reset(); }}>
              Сбросить к стандартным
            </button>
            <button type="button" disabled={saveMutation.isPending || !isValid || !changed}
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
          <p>Изменения видны до сохранения.</p>
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
    </main>
  );
}
