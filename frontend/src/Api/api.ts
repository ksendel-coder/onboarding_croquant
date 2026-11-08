import axios from "axios";
import { api, AUTH_EXPIRED_EVENT, AUTH_TOKEN_KEY } from "../Utils/constants";

api.interceptors.request.use((config) => {
  const token = sessionStorage.getItem(AUTH_TOKEN_KEY);
  if (token) config.headers.set("Authorization", `Bearer ${token}`);
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const url = error.config?.url ?? "";
      const isLoginOrRegister = url.includes("/auth/login") || url.includes("/auth/register");

      if (!isLoginOrRegister && sessionStorage.getItem(AUTH_TOKEN_KEY)) {
        sessionStorage.removeItem(AUTH_TOKEN_KEY);
        window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT));
      }
    }
    return Promise.reject(error);
  },
);

export { api };
