import { Navigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import { AuthBackground } from "./AuthBackground";
import { AuthPromo } from "./AuthPromo";
import { AuthForm } from "./AuthForm";
import styles from "./AuthPage.module.scss";

export type AuthMode = "login" | "register";

export function AuthPage({ mode }: { mode: AuthMode }) {
  const { status } = useAuth();

  if (status === "authenticated") return <Navigate to="/admin/scenarios" replace />;
  if (status === "checking") return <p>Проверяем авторизацию...</p>;

  return (
    <main className={styles.auth}>
      <div className={styles.auth__canvas}>
        <AuthBackground />
        <AuthPromo />
        <AuthForm mode={mode} />
      </div>
    </main>
  );
}
