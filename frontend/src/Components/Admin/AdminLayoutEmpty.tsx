import { memo } from "react";
import { Outlet } from "react-router-dom";
import { PageBackground } from "../UI/PageBackground/PageBackground";
import styles from "./Admin.module.scss";

function AdminLayoutEmptyComponent() {
  return (
    <div className={styles.layout}>
      <PageBackground />
      <div className={styles.layout__content}>
        <Outlet />
      </div>
    </div>
  );
}

export const AdminLayoutEmpty = memo(AdminLayoutEmptyComponent);