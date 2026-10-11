import { memo } from "react";
import { Button } from "../../UI/Button";
import styles from "./Styles.module.scss";
import { NavLink } from "react-router-dom";

function AdminTabsComponent() {
  return (
    <nav className={styles.tabs}>
      <NavLink to={"/admin/scenarios"}>
        {({ isActive }) => (
          <Button size="nav" color={isActive ? "primary" : "transparent"}>Сценарии</Button>
        )}
      </NavLink>
      <NavLink to={"/admin/analytics"}>
        {({ isActive }) => (
          <Button size="nav" color={isActive ? "primary" : "transparent"}>Аналитика</Button>
        )}
      </NavLink>
    </nav>
  );
}

export const AdminTabs = memo(AdminTabsComponent);
