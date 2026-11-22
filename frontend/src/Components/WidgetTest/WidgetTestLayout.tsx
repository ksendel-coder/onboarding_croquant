import { useEffect } from "react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { API_URL, ONBOARDING_APP_KEY } from "../../Config/env";
import styles from "./WidgetTest.module.scss";

export function WidgetTestLayout() {
  const { pathname } = useLocation();

  useEffect(() => {
    if (!ONBOARDING_APP_KEY) return;

    let active = true;
    let destroy: (() => void) | undefined;
    // Use an absolute URL so Vite serves the public module without an import query.
    const scriptUrl = new URL(`${import.meta.env.BASE_URL}onboarding-widget.js`, document.baseURI).href;
    void import(/* @vite-ignore */ scriptUrl)
      .then(({ mountOnboardingWidget }) => {
        if (active) destroy = mountOnboardingWidget({ appKey: ONBOARDING_APP_KEY, apiUrl: API_URL });
      })
      .catch((error: unknown) => {
        if (active) console.error("Не удалось подключить виджет онбординга.", error);
      });

    return () => {
      active = false;
      destroy?.();
    };
  }, [pathname]);

  return (
    <div className={styles.site}>
      <header className={styles.header}>
        <Link to="/widget-test" className={styles.logo}>Demo Site</Link>
        <nav className={styles.nav} aria-label="Меню сайта">
          <NavLink to="/widget-test" end>Главная</NavLink>
          <NavLink to="/widget-test/profile">Профиль</NavLink>
        </nav>
      </header>

      <Outlet />

      {!ONBOARDING_APP_KEY && (
        <p className={styles.error} role="alert">
          Задай VITE_ONBOARDING_APP_KEY в frontend/.env.local
        </p>
      )}
    </div>
  );
}
