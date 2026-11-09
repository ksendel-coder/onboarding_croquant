import styles from "./PageBackground.module.scss";

export function PageBackground() {
  return (
    <div className={styles.background} aria-hidden="true">
      <div className={styles.background__blurOne} />
      <div className={styles.background__blurTwo} />
      <div className={styles.background__blurThree} />
      <div className={styles.background__blurFour} />
    </div>
  );
}