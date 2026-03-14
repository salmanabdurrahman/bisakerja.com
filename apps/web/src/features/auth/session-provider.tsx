"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

import {
  hasBrowserSession,
  subscribeSessionChanged,
} from "@/lib/auth/session-cookie";

type AuthSessionState = "anonymous" | "authenticated";

interface AuthSessionContextValue {
  state: AuthSessionState;
  markAuthenticated: () => void;
  markAnonymous: () => void;
}

const AuthSessionContext = createContext<AuthSessionContextValue | null>(null);

export function AuthSessionProvider({ children }: PropsWithChildren) {
  const [state, setState] = useState<AuthSessionState>(() =>
    hasBrowserSession() ? "authenticated" : "anonymous",
  );

  useEffect(() => {
    return subscribeSessionChanged(() => {
      setState(hasBrowserSession() ? "authenticated" : "anonymous");
    });
  }, []);

  const markAuthenticated = useCallback(() => {
    setState("authenticated");
  }, []);

  const markAnonymous = useCallback(() => {
    setState("anonymous");
  }, []);

  const value = useMemo<AuthSessionContextValue>(
    () => ({
      state,
      markAuthenticated,
      markAnonymous,
    }),
    [state, markAuthenticated, markAnonymous],
  );

  return (
    <AuthSessionContext.Provider value={value}>
      {children}
    </AuthSessionContext.Provider>
  );
}

export function useAuthSession(): AuthSessionContextValue {
  const context = useContext(AuthSessionContext);
  if (!context) {
    throw new Error("useAuthSession must be used within AuthSessionProvider");
  }
  return context;
}
