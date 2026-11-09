import { memo, useState } from "react";
import logo from "/logo.svg";
import styles from "./Styles.module.scss";
import { useAuth } from "../../Auth/AuthContext";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";

function HeaderComponent() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [busy, setBusy] = useState(false);

  const handleLogout = async () => {
    setBusy(true);
    try {
      await logout();
    } finally {
      queryClient.clear();
      navigate("/login", { replace: true });
      setBusy(false);
    }
  };

  return (
    <header className={styles.header}>
      <div className={styles.header__brand}>
        <img src={logo} alt="Логотип" className={styles.header__brand__logo} />
        <h2 className={styles.header__brand__title}>
          Onboarding{" "}
          <span className={styles.header__brand__title__span}>Studio</span>
        </h2>
      </div>
      <div className={styles.header__account}>
        <button type="button" className={styles.header__profile} aria-label="Меню профиля" aria-haspopup="true">
          <span className={styles.header__avatar}>{user?.email?.charAt(0).toUpperCase() ?? "A"}</span>
        </button>

        <div className={styles.header__dropdown}>
          <Link to="/admin/profile" className={styles.header__dropdownItem}>Профиль</Link>
          <button type="button" className={styles.header__dropdownItem} onClick={() => void handleLogout()} disabled={busy}>
            Выйти
          </button>
        </div>
      </div>
    </header>
  );
}

export const Header = memo(HeaderComponent);
