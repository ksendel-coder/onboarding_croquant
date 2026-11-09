import logo from "/logo.svg";
import { Checkbox } from "../UI/Checkbox";
import styles from "./AuthPage.module.scss";

const benefits = ["Интерактивные сценарии", "Гибкая аналитика", "Быстрая интеграция"];

export function AuthPromo() {
  return (
    <section className={styles.auth__promo}>
      <img className={styles.auth__logo} src={logo} alt="" />
      <h1 className={styles.auth__brand}>Onboarding <span>Studio</span></h1>
      <p className={styles.auth__tagline}>Создавайте онбординг,<br />который работает</p>
      <ul className={styles.auth__benefits}>
        {benefits.map((benefit) => (
          <li key={benefit}>
            <Checkbox checked={true} onChange={() => {}} />
            {benefit}
          </li>
        ))}
      </ul>
    </section>
  );
}
