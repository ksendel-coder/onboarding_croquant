import { BrowserRouter, Routes, Route, Outlet, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AdminLayoutWithTabs } from "../Components/Admin/AdminLayoutWithTabs";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { Header } from "../Components/Layouts/Header/Header";
import { ScenarioEditor } from "../Components/Admin/Pages/ScenarioEditor/ScenarioEditor";
import { AdminLayoutEmpty } from "../Components/Admin/AdminLayoutEmpty";
import { Analytics } from "../Components/Admin/Pages/Analytics/Analytics";
import { AnalyticsDetail } from "../Components/Admin/Pages/Analytics/AnalyticsDetail";

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/admin" element={<WrappPages />}>
            <Route element={<AdminLayoutWithTabs />}>
              <Route index element={<Navigate to="scenarios" replace />} />
              <Route path="scenarios" element={<Scenarios />} />
              <Route path="analytics" element={<Analytics />} />
              <Route path="analytics/:tourId" element={<AnalyticsDetail />} />
            </Route>
            <Route element={<AdminLayoutEmpty />}>
              <Route path="scenarios/new" element={<ScenarioEditor />} />
              <Route path="scenarios/:tourId/edit" element={<ScenarioEditor />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

function WrappPages() {
  return (
    <>
      <Header />
      <Outlet />
    </>
  );
}

export default App;
