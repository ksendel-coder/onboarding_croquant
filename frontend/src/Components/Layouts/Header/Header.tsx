import { memo } from "react";
import logo from "../../../../public/logo.svg";
import styles from './Styles.module.scss';

function HeaderComponent() {
  return (
    <header className={styles.header}>
      <div className={styles.header__brand}>
        <img src={logo} alt="Логотип" className={styles.header__brand__logo}/>
        <h2 className={styles.header__brand__title}>Onboarding <span className={styles.header__brand__title__span}>Studio</span></h2>
      </div>
      <div className={styles.header__profile}>ИТ</div>
    </header>
  );
}

export const Header = memo(HeaderComponent);
