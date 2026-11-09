import { useState, type FormEvent } from "react";
import axios from "axios";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import { PasswordField } from "./PasswordField";
import type { AuthMode } from "./AuthPage";
import styles from "./AuthPage.module.scss";

export function AuthForm({ mode }: { mode: AuthMode }) {
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const isRegister = mode === "register";
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    if (isRegister && (!firstName.trim() || !lastName.trim())) {
      setError("Введите имя и фамилию.");
      return;
    }
    if (!email.trim() || !password) {
      setError("Введите email и пароль.");
      return;
    }
    if (isRegister && password.length < 8) {
      setError("Пароль должен содержать не менее 8 символов.");
      return;
    }
    if (isRegister && password !== confirmPassword) {
      setError("Пароли не совпадают.");
      return;
    }
    setBusy(true);

    try {
      const credentials = { email: email.trim(), password };
      if (isRegister) await register(credentials);
      else await login(credentials);

      const from = (location.state as { from?: unknown } | null)?.from;
      navigate(typeof from === "string" && from.startsWith("/admin") ? from : "/admin/scenarios", { replace: true });

    } catch (err) {
      if (axios.isAxiosError(err)) {
        const code = err.response?.status;
        if (code === 401) setError("Неверный email или пароль.");
        else if (code === 409) setError("Пользователь с таким email уже существует.");
        else if (code === 400) setError("Проверьте правильность введённых данных.");
        else setError("Сервер недоступен. Попробуйте ещё раз.");
      } else {
        setError("Не удалось выполнить запрос.");
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className={styles.auth__card} aria-label={isRegister ? "Регистрация" : "Вход"}>
      <h2>{isRegister ? "Регистрация" : "Вход"}</h2>
      <form className={styles.auth__form} onSubmit={(event) => void onSubmit(event)}>
        {isRegister && (
          <>
            <label htmlFor="auth-first-name">Имя</label>
            <input
              id="auth-first-name"
              autoComplete="given-name"
              required
              value={firstName}
              onChange={(event) => setFirstName(event.target.value)}
            />
            <label htmlFor="auth-last-name">Фамилия</label>
            <input
              id="auth-last-name"
              autoComplete="family-name"
              required
              value={lastName}
              onChange={(event) => setLastName(event.target.value)}
            />
          </>
        )}
        <label htmlFor="auth-email">Email</label>
        <input
          id="auth-email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />

        <PasswordField
          key={`password-${mode}`}
          id="auth-password"
          label="Пароль"
          value={password}
          onChange={setPassword}
          autoComplete={isRegister ? "new-password" : "current-password"}
          minLength={isRegister ? 8 : undefined}
        />

        {isRegister && (
          <PasswordField
            id="auth-confirm-password"
            label="Повторите пароль"
            value={confirmPassword}
            onChange={setConfirmPassword}
            autoComplete="new-password"
            minLength={8}
          />
        )}

        {error && <p className={styles.auth__error} role="alert">{error}</p>}
        <button className={styles.auth__submit} type="submit" disabled={busy}>
          {busy ? "Загрузка..." : isRegister ? "Зарегистрироваться" : "Войти в кабинет"}
        </button>
      </form>

      <p className={styles.auth__footer}>
        {isRegister ? (<>Уже есть профиль? <Link to="/login">Войти</Link></>) : (
          <>Еще нет профиля? <Link to="/register">Регистрация</Link></>
        )}
      </p>
    </section>
  );
}
