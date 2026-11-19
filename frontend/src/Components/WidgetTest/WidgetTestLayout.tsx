import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { OnboardingProvider } from "../Onboarding/OnboardingProvider";
import { API_URL, ONBOARDING_APP_KEY } from "../../Config/env";
import styles from "./WidgetTest.module.scss";

export function WidgetTestLayout() {
  const { pathname } = useLocation();

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

      {ONBOARDING_APP_KEY ? (
        <OnboardingProvider
          key={pathname}
          appKey={ONBOARDING_APP_KEY}
          apiUrl={API_URL.replace(/\/$/, "")}
        />
      ) : (
        <p className={styles.error} role="alert">
          Задай VITE_ONBOARDING_APP_KEY в frontend/.env.local
        </p>
      )}
    </div>
  );
}
