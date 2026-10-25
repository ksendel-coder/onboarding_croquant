import { memo } from "react";
import { Outlet } from "react-router-dom";

function AdminLayoutEmptyComponent() {
  return (
    <div>
      <Outlet />
    </div>
  );
}

export const AdminLayoutEmpty = memo(AdminLayoutEmptyComponent);