import { useState, type FormEvent } from "react";
import axios from "axios";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import logo from "/logo.svg";
import styles from "./AuthPage.module.scss";
import { Checkbox } from "../UI/Checkbox";

type Mode = "login" | "register";

function EyeIcon({ visible }: { visible: boolean }) {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden="true">
      <path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6S2 12 2 12Z" />
      <circle cx="12" cy="12" r="3" />
      {!visible && <path d="M3 21 21 3" />}
    </svg>
  );
}

export function AuthPage({ mode }: { mode: Mode }) {
  const { status, login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const isRegister = mode === "register";

  if (status === "authenticated") return <Navigate to="/admin/scenarios" replace />;
  if (status === "checking") return <p>Проверяем авторизацию...</p>;

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
        if (err.response?.status === 401) setError("Неверный email или пароль.");
        else if (err.response?.status === 409) setError("Пользователь с таким email уже существует.");
        else if (err.response?.status === 400) setError("Проверьте правильность введённых данных.");
        else setError("Сервер недоступен. Попробуйте ещё раз.");
      } else {
        setError("Не удалось выполнить запрос.");
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className={styles.auth}>
      <div className={styles.auth__canvas}>
        <div className={styles.auth__blurOne} aria-hidden="true" />
        <div className={styles.auth__blurTwo} aria-hidden="true" />
        <div className={styles.auth__blurThree} aria-hidden="true" />
        <div className={styles.auth__blurFour} aria-hidden="true" />
        <div className={styles.auth__orbit} aria-hidden="true" />
        <div className={styles.auth__orbitBig} aria-hidden="true" />
        <div className={`${styles.auth__float} ${styles.auth__float_top}`} aria-hidden="true">▥</div>
        <div className={`${styles.auth__float} ${styles.auth__float_middle}`} aria-hidden="true">◴</div>
        <div className={`${styles.auth__float} ${styles.auth__float_bottom}`} aria-hidden="true">▤</div>
        <span className={styles.auth__dot} aria-hidden="true" />

        <section className={styles.auth__promo} aria-label="О продукте">
          <img className={styles.auth__logo} src={logo} />
          <h1 className={styles.auth__brand}>Onboarding <span>Studio</span></h1>
          <p className={styles.auth__tagline}>Создавайте онбординг,<br />который работает</p>
          <ul className={styles.auth__benefits}>
            <li> <Checkbox checked={true} onChange={() => {}} />Интерактивные сценарии</li>
            <li> <Checkbox checked={true} onChange={() => {}} />Гибкая аналитика</li>
            <li> <Checkbox checked={true} onChange={() => {}} />Быстрая интеграция</li>
          </ul>
        </section>

        <section className={styles.auth__card} aria-label={isRegister ? "Регистрация" : "Вход"}>
          <h2>{isRegister ? <>Регистрация</> : "Вход"}</h2>

          <form className={styles.auth__form} onSubmit={(event) => { void onSubmit(event); }}>
            {isRegister && (
              <>
                <label htmlFor="auth-first-name">Имя</label>
                <input id="auth-first-name" autoComplete="given-name" required value={firstName}
                  onChange={(event) => setFirstName(event.target.value)} />
                <label htmlFor="auth-last-name">Фамилия</label>
                <input id="auth-last-name" autoComplete="family-name" required value={lastName}
                  onChange={(event) => setLastName(event.target.value)} />
              </>
            )}
            <label htmlFor="auth-email">Email</label>
            <input id="auth-email" type="email" autoComplete="email" required value={email}
              onChange={(event) => setEmail(event.target.value)} />

            <label htmlFor="auth-password">Пароль</label>
            <div className={styles.auth__passwordField}>
              <input id="auth-password" type={showPassword ? "text" : "password"} required
                autoComplete={isRegister ? "new-password" : "current-password"}
                minLength={isRegister ? 8 : undefined} value={password}
                onChange={(event) => setPassword(event.target.value)} />
              <button type="button" aria-label={showPassword ? "Скрыть пароль" : "Показать пароль"}
                aria-pressed={showPassword} onClick={() => setShowPassword((value) => !value)}>
                <EyeIcon visible={showPassword} />
              </button>
            </div>

            {isRegister && (
              <>
                <label htmlFor="auth-confirm-password">Повторите пароль</label>
                <div className={styles.auth__passwordField}>
                  <input id="auth-confirm-password" type={showConfirmPassword ? "text" : "password"}
                    required autoComplete="new-password" minLength={8} value={confirmPassword}
                    onChange={(event) => setConfirmPassword(event.target.value)} />
                  <button type="button" aria-label={showConfirmPassword ? "Скрыть повтор пароля" : "Показать повтор пароля"}
                    aria-pressed={showConfirmPassword}
                    onClick={() => setShowConfirmPassword((value) => !value)}>
                    <EyeIcon visible={showConfirmPassword} />
                  </button>
                </div>
              </>
            )}

            {error && <p className={styles.auth__error} role="alert">{error}</p>}
            <button className={styles.auth__submit} type="submit" disabled={busy}>
              {busy ? "Отправка..." : isRegister ? "Зарегистрироваться" : "Войти в кабинет"}
            </button>
          </form>
          <p className={styles.auth__footer}>
            {isRegister ? <>Уже есть профиль? <Link to="/login">Войти</Link></>
              : <>Еще нет профиля? <Link to="/register">Регистрация</Link></>}
          </p>
        </section>
      </div>
    </main>
  );
}