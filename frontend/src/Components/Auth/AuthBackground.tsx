import diagram from "/diagram.svg";
import clock from "/clock.svg";
import money from "/money.svg";
import styles from "./AuthPage.module.scss";

export function AuthBackground() {
  return (
    <>
      <div className={styles.auth__blurOne} aria-hidden="true" />
      <div className={styles.auth__blurTwo} aria-hidden="true" />
      <div className={styles.auth__blurThree} aria-hidden="true" />
      <div className={styles.auth__blurFour} aria-hidden="true" />
      <div className={styles.auth__orbit} aria-hidden="true" />
      <div className={styles.auth__orbitBig} aria-hidden="true" />

      <div className={`${styles.auth__float} ${styles.auth__float_top}`} aria-hidden="true">
        <img src={diagram} alt="" />
      </div>
      <div className={`${styles.auth__float} ${styles.auth__float_middle}`} aria-hidden="true">
        <img src={clock} alt="" />
      </div>
      <div className={`${styles.auth__float} ${styles.auth__float_bottom}`} aria-hidden="true">
        <img src={money} alt="" />
      </div>
      <span className={styles.auth__dot} aria-hidden="true" />
    </>
  );
}
