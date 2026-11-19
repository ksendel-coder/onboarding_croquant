import { useState } from "react";
import { Link } from "react-router-dom";
import styles from "./WidgetTest.module.scss";

export function WidgetTestHome() {
  const [message, setMessage] = useState("");

  return (
    <main className={styles.main}>
      <h1 id="home-title">Добро пожаловать!</h1>
      <p>Это простой сайт для проверки настоящего виджета онбординга.</p>

      <section id="welcome-card" className={styles.card}>
        <h2>Начало работы</h2>
        <p>Нажми кнопку и проверь расположение подсказки.</p>
        <button
          id="start-button"
          type="button"
          onClick={() => setMessage("Кнопка работает!")}
        >
          Начать
        </button>
        {message && <p role="status">{message}</p>}
      </section>

      <Link id="go-profile" to="/widget-test/profile" className={styles.actionLink}>
        Перейти в профиль →
      </Link>
    </main>
  );
}
