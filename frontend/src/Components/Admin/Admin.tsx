import { memo } from "react";
import { Outlet } from "react-router-dom";
import { AdminTabs } from "./AdminTabs/AdminTabs";

function AdminComponent() {
  return (
    <div>
      <AdminTabs />
      <Outlet />
    </div>
  );
}

export const Admin = memo(AdminComponent);