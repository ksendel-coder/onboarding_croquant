export interface AuthUser {
  id: string;
  email: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}

export interface Credentials {
  email: string;
  password: string;
}

export type AuthStatus = "checking" | "authenticated" | "anonymous";

export interface AuthContextValue {
  status: AuthStatus;
  user: AuthUser | null;
  login: (credentials: Credentials) => Promise<void>;
  register: (credentials: Credentials) => Promise<void>;
  logout: () => Promise<void>;
}
