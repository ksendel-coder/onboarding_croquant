import { BrowserRouter, Routes, Route, Outlet, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AdminLayoutWithTabs } from "../Components/Admin/AdminLayoutWithTabs";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { Header } from "../Components/Layouts/Header/Header";
import { ScenarioEditor } from "../Components/Admin/Pages/ScenarioEditor/ScenarioEditor";
import { AdminLayoutEmpty } from "../Components/Admin/AdminLayoutEmpty";
import { Analytics } from "../Components/Admin/Pages/Analytics/Analytics";
import { AnalyticsDetail } from "../Components/Admin/Pages/Analytics/AnalyticsDetail";
import { AuthProvider } from "../Components/Auth/AuthProvider";
import { AuthPage } from "../Components/Auth/AuthPage";
import { ProtectedRoute } from "../Components/Auth/ProtectedRoute";
import { ScenarioPreview } from "../Components/Admin/Pages/ScenarioPreview/ScenarioPreview";
import { ThemeSettings } from "../Components/Admin/Pages/ThemeSettings/ThemeSettings";

import { WidgetTestLayout } from "../Components/WidgetTest/WidgetTestLayout";
import { WidgetTestHome } from "../Components/WidgetTest/WidgetTestHome";
import { WidgetTestProfile } from "../Components/WidgetTest/WidgetTestProfile";

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route path="/" element={<Navigate to="/admin" replace />} />
            <Route path="/login" element={<AuthPage mode="login" />} />
            <Route path="/register" element={<AuthPage mode="register" />} />

            <Route path="/widget-test" element={<WidgetTestLayout />}>
              <Route index element={<WidgetTestHome />} />
              <Route path="profile" element={<WidgetTestProfile />} />
            </Route>

            <Route element={<ProtectedRoute />}>
              <Route path="/admin" element={<AdminShell />}>
                <Route element={<AdminLayoutWithTabs />}>
                  <Route index element={<Navigate to="scenarios" replace />} />
                  <Route path="scenarios" element={<Scenarios />} />
                  <Route path="analytics" element={<Analytics />} />
                  <Route path="theme" element={<ThemeSettings />} />
                </Route>
                <Route element={<AdminLayoutEmpty />}>
                  <Route path="scenarios/new" element={<ScenarioEditor />} />
                  <Route path="scenarios/:tourId/edit" element={<ScenarioEditor />} />
                  <Route path="scenarios/:tourId/preview" element={<ScenarioPreview />} />
                  <Route path="analytics/:tourId" element={<AnalyticsDetail />} />
                </Route>
              </Route>
            </Route>
            <Route path="*" element={<Navigate to="/admin" replace />} />
          </Routes>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

function AdminShell() {
  return <><Header /><Outlet /></>;
}
