import { useState } from "react";
import styles from "./AuthPage.module.scss";

type PasswordFieldProps = {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete: "current-password" | "new-password";
  minLength?: number;
};

function EyeIcon({ visible }: { visible: boolean }) {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden="true">
      <path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6S2 12 2 12Z" />
      <circle cx="12" cy="12" r="3" />
      {!visible && <path d="M3 21 21 3" />}
    </svg>
  );
}

export function PasswordField({ id, label, value, onChange, autoComplete, minLength }: PasswordFieldProps) {
  const [visible, setVisible] = useState(false);

  return (
    <>
      <label htmlFor={id}>{label}</label>
      <div className={styles.auth__passwordField}>
        <input
          id={id}
          type={visible ? "text" : "password"}
          required
          autoComplete={autoComplete}
          minLength={minLength}
          value={value}
          onChange={(event) => onChange(event.target.value)}
        />
        <button
          type="button"
          aria-label={visible ? "Скрыть пароль" : "Показать пароль"}
          aria-pressed={visible}
          onClick={() => setVisible((current) => !current)}
        >
          <EyeIcon visible={visible} />
        </button>
      </div>
    </>
  );
}
