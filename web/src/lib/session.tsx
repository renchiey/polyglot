import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { ApiError, auth, createApi, type Api } from "./api";

// Auth UI is intentionally not built yet (per the roadmap). To still exercise
// the authenticated endpoints, the app silently signs in a fixed dev account on
// load — logging in, or registering it the first time. Replace this with the
// real auth flow in a later phase; nothing else depends on it.
const DEV_EMAIL = "dev@context.app";
const DEV_PASSWORD = "context-dev-001";

// The API base URL is configured at build time via .env (VITE_API_URL); there
// is no in-app override.
const API_BASE = import.meta.env.VITE_API_URL?.replace(/\/$/, "") || "http://localhost:8080";

type Status = "loading" | "ready" | "error";

interface SessionValue {
  api: Api | null;
  status: Status;
  error: string | null;
  retry: () => void;
}

const SessionContext = createContext<SessionValue | null>(null);

async function bootstrapToken(): Promise<string> {
  try {
    const res = await auth.login(API_BASE, DEV_EMAIL, DEV_PASSWORD);
    return res.token;
  } catch (err) {
    // 401 = account doesn't exist yet; create it. Anything else re-throws.
    if (err instanceof ApiError && err.status === 401) {
      const res = await auth.register(API_BASE, DEV_EMAIL, DEV_PASSWORD);
      return res.token;
    }
    throw err;
  }
}

export function SessionProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [status, setStatus] = useState<Status>("loading");
  const [error, setError] = useState<string | null>(null);
  const attempt = useRef(0);

  const connect = useCallback(() => {
    const id = ++attempt.current;
    setStatus("loading");
    setError(null);
    setToken(null);
    bootstrapToken()
      .then((t) => {
        if (id === attempt.current) {
          setToken(t);
          setStatus("ready");
        }
      })
      .catch((err: unknown) => {
        if (id !== attempt.current) return;
        setStatus("error");
        setError(err instanceof Error ? err.message : "Could not connect");
      });
  }, []);

  useEffect(connect, [connect]);

  const value: SessionValue = {
    api: token ? createApi(API_BASE, token) : null,
    status,
    error,
    retry: connect,
  };

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionValue {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within SessionProvider");
  return ctx;
}

/** Convenience for views that only run once the session is ready. */
export function useApi(): Api {
  const { api } = useSession();
  if (!api) throw new Error("API used before session ready");
  return api;
}
