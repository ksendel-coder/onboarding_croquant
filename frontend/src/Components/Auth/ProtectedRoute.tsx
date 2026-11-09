import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "./AuthContext";

export function ProtectedRoute() {
  const { status } = useAuth();
  const location = useLocation();

  if (status === "checking") return <p>Проверяем авторизацию...</p>;
  if (status !== "authenticated") return <Navigate to="/login" replace state={{ from: location.pathname }} />;

  return <Outlet />;
}