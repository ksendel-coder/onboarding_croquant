import type { AuthResponse, AuthUser, Credentials } from "../Types";
import { api } from "./api";

export const authAPI = {
  register: (credentials: Credentials) => api.post<AuthResponse>("/auth/register", credentials).then((r) => r.data),

  login: (credentials: Credentials) => api.post<AuthResponse>("/auth/login", credentials).then((r) => r.data),

  me: () => api.get<AuthUser>("/auth/me").then((r) => r.data),

  logout: () => api.post<void>("/auth/logout").then(() => undefined)
};