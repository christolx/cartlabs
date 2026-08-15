"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import type { components } from "@/lib/api/schema";
import { BrowserAPIError, browserRequest } from "@/lib/api/browser";

type Session = components["schemas"]["Session"];
type User = components["schemas"]["User"];
type Role = components["schemas"]["Role"];
type SessionStatus = "loading" | "anonymous" | "authenticated";

type SessionContextValue = {
  demoEnabled: boolean;
  status: SessionStatus;
  user: User | null;
  login: (email: string, password: string) => Promise<User>;
  demoLogin: (role: Role) => Promise<User>;
  logout: () => Promise<void>;
  request: <T>(path: string, init?: RequestInit) => Promise<T>;
};

const SessionContext = createContext<SessionContextValue | null>(null);

export function actorHome(role: Role) {
  if (role === "seller") return "/seller";
  if (role === "admin") return "/admin";
  return "/";
}

export function SessionProvider({ children, demoEnabled }: { children: ReactNode; demoEnabled: boolean }) {
  const [status, setStatus] = useState<SessionStatus>("loading");
  const [user, setUser] = useState<User | null>(null);
  const tokenRef = useRef("");
  const refreshRef = useRef<Promise<string> | null>(null);

  const clearSession = useCallback(() => {
    tokenRef.current = "";
    setUser(null);
    setStatus("anonymous");
  }, []);

  const acceptSession = useCallback(async (session: Session) => {
    tokenRef.current = session.accessToken;
    const currentUser = await browserRequest<User>("/me", session.accessToken);
    setUser(currentUser);
    setStatus("authenticated");
    return currentUser;
  }, []);

  const refresh = useCallback(async () => {
    if (refreshRef.current) return refreshRef.current;
    refreshRef.current = (async () => {
      try {
        const session = await browserRequest<Session>("/auth/refresh", undefined, { method: "POST" });
        await acceptSession(session);
        return session.accessToken;
      } catch (error) {
        clearSession();
        throw error;
      } finally {
        refreshRef.current = null;
      }
    })();
    return refreshRef.current;
  }, [acceptSession, clearSession]);

  useEffect(() => {
    void refresh().catch(() => undefined);
  }, [refresh]);

  const request = useCallback(async <T,>(path: string, init?: RequestInit) => {
    try {
      return await browserRequest<T>(path, tokenRef.current, init);
    } catch (error) {
      if (!(error instanceof BrowserAPIError) || error.status !== 401 || path.startsWith("/auth/")) throw error;
      const token = await refresh();
      return browserRequest<T>(path, token, init);
    }
  }, [refresh]);

  const login = useCallback(async (email: string, password: string) => {
    const session = await browserRequest<Session>("/auth/login", undefined, {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    return acceptSession(session);
  }, [acceptSession]);

  const demoLogin = useCallback(async (role: Role) => {
    const session = await browserRequest<Session>("/auth/demo-login", undefined, {
      method: "POST",
      body: JSON.stringify({ role }),
    });
    return acceptSession(session);
  }, [acceptSession]);

  const logout = useCallback(async () => {
    try {
      await browserRequest<void>("/auth/logout", tokenRef.current, { method: "POST" });
    } finally {
      clearSession();
    }
  }, [clearSession]);

  const value = useMemo<SessionContextValue>(() => ({
    demoEnabled,
    status,
    user,
    login,
    demoLogin,
    logout,
    request,
  }), [demoEnabled, status, user, login, demoLogin, logout, request]);

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession() {
  const context = useContext(SessionContext);
  if (!context) throw new Error("useSession must be used inside SessionProvider");
  return context;
}
