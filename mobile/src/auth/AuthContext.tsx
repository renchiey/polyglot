import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import * as api from "@/api/auth";
import { tokenStorage } from "@/lib/storage";

const TOKEN_KEY = "auth.token";

type AuthState = {
  user: api.User | null;
  token: string | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
};

const AuthContext = createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<api.User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // On mount, restore a saved token and fetch the current user.
  useEffect(() => {
    (async () => {
      try {
        const saved = await tokenStorage.get(TOKEN_KEY);
        if (saved) {
          const profile = await api.me(saved);
          setToken(saved);
          setUser(profile);
        }
      } catch {
        await tokenStorage.remove(TOKEN_KEY);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const apply = useCallback(async (res: api.AuthResponse) => {
    await tokenStorage.set(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  const signIn = useCallback(
    async (email: string, password: string) => apply(await api.login(email, password)),
    [apply],
  );

  const signUp = useCallback(
    async (email: string, password: string) => apply(await api.register(email, password)),
    [apply],
  );

  const signOut = useCallback(async () => {
    await tokenStorage.remove(TOKEN_KEY);
    setToken(null);
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, token, loading, signIn, signUp, signOut }),
    [user, token, loading, signIn, signUp, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within an AuthProvider");
  return ctx;
}
