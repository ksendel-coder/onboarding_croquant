import { useState } from "react";
import styles from "./WidgetTest.module.scss";

export function WidgetTestProfile() {
  const [name, setName] = useState("");
  const [saved, setSaved] = useState(false);

  return (
    <main className={styles.main}>
      <h1 id="profile-title">Мой профиль</h1>
      <section id="profile-card" className={styles.card}>
        <label htmlFor="profile-name">Имя</label>
        <input
          id="profile-name"
          value={name}
          onChange={(event) => {
            setName(event.target.value);
            setSaved(false);
          }}
          placeholder="Введите имя"
        />
        <button
          id="save-profile"
          type="button"
          onClick={() => setSaved(true)}
        >
          Сохранить
        </button>
        {saved && <p role="status">Сохранено для проверки интерфейса</p>}
      </section>
    </main>
  );
}
