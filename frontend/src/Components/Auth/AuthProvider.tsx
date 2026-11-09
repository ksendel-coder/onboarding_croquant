import { useEffect, useState, type ReactNode } from "react";
import { authAPI } from "../../Api/auth";
import { AuthContext } from "./AuthContext";
import type { AuthStatus, AuthUser, Credentials } from "../../Types";
import { AUTH_EXPIRED_EVENT, AUTH_TOKEN_KEY } from "../../Utils/constants";
import { useQueryClient } from "@tanstack/react-query";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [status, setStatus] = useState<AuthStatus>(() => sessionStorage.getItem(AUTH_TOKEN_KEY) ? "checking" : "anonymous");
  const [checkError, setCheckError] = useState(false);
  const [checkNumber, setCheckNumber] = useState(0);
  const queryClient = useQueryClient();

  useEffect(() => {
    const onExpired = () => {
      queryClient.clear();
      setUser(null);
      setStatus("anonymous");
    };
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired);
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired);
  }, [queryClient]);

  useEffect(() => {
    const token = sessionStorage.getItem(AUTH_TOKEN_KEY);
    if (!token) return;

    let active = true;
    authAPI
      .me()
      .then((currentUser) => {
        if (!active) return;

        setUser(currentUser);
        setCheckError(false);
        setStatus("authenticated");
      })
      .catch(() => {
        if (!active) return;

        if (!sessionStorage.getItem(AUTH_TOKEN_KEY)) {
          setUser(null);
          setStatus("anonymous");
        } else {
          setCheckError(true);
          setStatus("checking");
        }
      });

    return () => {
      active = false;
    };
  }, [checkNumber]);

  const acceptAuth = (token: string, currentUser: AuthUser) => {
    sessionStorage.setItem(AUTH_TOKEN_KEY, token);
    setUser(currentUser);
    setCheckError(false);
    setStatus("authenticated");
  };

  const login = async (credentials: Credentials) => {
    const response = await authAPI.login(credentials);
    acceptAuth(response.token, response.user);
  };

  const register = async (credentials: Credentials) => {
    const response = await authAPI.register(credentials);
    acceptAuth(response.token, response.user);
  };

  const logout = async () => {
    try {
      await authAPI.logout();
    } finally {
      sessionStorage.removeItem(AUTH_TOKEN_KEY);
      setUser(null);
      setStatus("anonymous");
    }
  };

  return (
    <AuthContext.Provider value={{ status, user, login, register, logout }}>
      {status === "checking" && checkError ? (
        <div role="alert" style={{ padding: 24 }}>
          <p>Не удалось проверить сессию. Проверьте подключение к серверу.</p>
          <button type="button" onClick={() => { setCheckError(false); setCheckNumber((n) => n + 1); }}>
            Повторить
          </button>
          <button
            type="button"
            onClick={() => {
              sessionStorage.removeItem(AUTH_TOKEN_KEY);
              setUser(null);
              setStatus("anonymous");
            }}
          >
            Войти заново
          </button>
        </div>
      ) : (children)}
    </AuthContext.Provider>
  );
}
