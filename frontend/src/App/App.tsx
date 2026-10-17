import { BrowserRouter, Routes, Route, Outlet, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Admin } from "../Components/Admin/Admin";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { Header } from "../Components/Layouts/Header/Header";

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/admin" element={<WrappPages />}>
            <Route element={<Admin />}>
              <Route index element={<Navigate to="scenarios" replace />} />
              <Route path="scenarios" element={<Scenarios />} />
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
