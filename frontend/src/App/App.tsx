import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Admin } from "../Components/Admin/Admin";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { Header } from "../Components/Layouts/Header/Header";

function App() {
  return (
    <QueryClientProvider client={new QueryClient()}>
      <BrowserRouter>
        <Routes>
          <Route path="/admin" element={<WrappPages />}>
            <Route index element={<Admin />} />
            <Route path="scenarios" element={<Scenarios />} />
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
