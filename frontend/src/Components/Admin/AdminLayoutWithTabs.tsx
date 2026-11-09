import { memo } from "react";
import { Outlet } from "react-router-dom";
import { AdminTabs } from "./AdminTabs/AdminTabs";
import { PageBackground } from "../UI/PageBackground/PageBackground";
import styles from "./Admin.module.scss";

function AdminLayoutWithTabsComponent() {

  return (
    <div className={styles.layout}>
      <PageBackground />
      <div className={styles.layout__content}>
        <AdminTabs />
        <Outlet />
      </div>
    </div>
  );
}

export const AdminLayoutWithTabs = memo(AdminLayoutWithTabsComponent);