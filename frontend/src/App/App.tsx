import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Admin } from "../Components/Admin/Admin";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";

function App() {
  return (
    <QueryClientProvider client={new QueryClient()}>
      <BrowserRouter>
        <Routes>
          <Route path="/admin" element={<Admin />}>
            <Route path="scenarios" element={<Scenarios />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
