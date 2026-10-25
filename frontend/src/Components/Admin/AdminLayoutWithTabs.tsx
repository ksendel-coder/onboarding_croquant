import { memo } from "react";
import { Outlet } from "react-router-dom";
import { AdminTabs } from "./AdminTabs/AdminTabs";

function AdminLayoutWithTabsComponent() {

  return (
    <div>
      <AdminTabs />
      <Outlet />
    </div>
  );
}

export const AdminLayoutWithTabs = memo(AdminLayoutWithTabsComponent);