import { memo } from "react";
import { Outlet } from "react-router-dom";

function AdminComponent() {
  return (
    <div>
      <Outlet />
    </div>
  );
}

export const Admin = memo(AdminComponent);