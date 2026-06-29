import {
  createContext,
  useCallback,
  useContext,
  useState,
  type ButtonHTMLAttributes,
  type CSSProperties,
  type ReactNode,
} from "react";
import { Check, Loader2, X } from "lucide-react";

// Accent palette used to colour-code features as distinct blocks. Deep,
// restrained tones (indigo spine) keep it adult/educational, not candy.
export const ACCENTS = {
  grape: "var(--color-grape)",
  bubble: "var(--color-bubble)",
  tangerine: "var(--color-tangerine)",
  sunshine: "var(--color-sunshine)",
  mint: "var(--color-mint)",
  sky: "var(--color-sky)",
  success: "var(--color-success)",
  danger: "var(--color-danger)",
} as const;
export type Accent = keyof typeof ACCENTS;

const cx = (...parts: (string | false | undefined)[]) => parts.filter(Boolean).join(" ");
const vars = (obj: Record<string, string>) => obj as CSSProperties;

// ── Button ──────────────────────────────────────────────────────────────
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  accent?: Accent;
  soft?: boolean;
  loading?: boolean;
}
export function Button({
  accent = "grape",
  soft = false,
  loading = false,
  className,
  children,
  disabled,
  style,
  ...rest
}: ButtonProps) {
  return (
    <button
      className={cx("clay-btn", soft && "clay-btn-soft", className)}
      style={{ ...vars({ "--btn": ACCENTS[accent] }), ...style }}
      disabled={disabled || loading}
      {...rest}
    >
      {loading && <Loader2 size={18} className="animate-spin" />}
      {children}
    </button>
  );
}

// ── Surfaces ────────────────────────────────────────────────────────────
export function Card({
  className,
  children,
  style,
  ...rest
}: { className?: string; children: ReactNode; style?: CSSProperties } & React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cx("clay p-5 sm:p-6", className)} style={style} {...rest}>
      {children}
    </div>
  );
}

export function TintPanel({
  accent = "grape",
  className,
  children,
  style,
}: {
  accent?: Accent;
  className?: string;
  children: ReactNode;
  style?: CSSProperties;
}) {
  return (
    <div
      className={cx("clay-tint p-5 sm:p-6", className)}
      style={{ ...vars({ "--tint": ACCENTS[accent] }), ...style }}
    >
      {children}
    </div>
  );
}

export function Pill({
  accent = "grape",
  children,
  className,
}: {
  accent?: Accent;
  children: ReactNode;
  className?: string;
}) {
  return (
    <span className={cx("pill", className)} style={vars({ "--tint": ACCENTS[accent] })}>
      {children}
    </span>
  );
}

export function Spinner({ size = 22 }: { size?: number }) {
  return <Loader2 size={size} className="animate-spin text-[var(--color-grape)]" />;
}

// ── Form controls ───────────────────────────────────────────────────────
export function Field({
  label,
  children,
  hint,
}: {
  label: string;
  children: ReactNode;
  hint?: string;
}) {
  return (
    <label className="block">
      <span className="mb-1.5 block font-display text-sm font-bold text-[var(--color-ink)]">
        {label}
      </span>
      {children}
      {hint && <span className="mt-1 block text-xs text-[var(--color-ink-soft)]">{hint}</span>}
    </label>
  );
}

const inputClass =
  "clay-inset w-full px-4 py-2.5 text-[var(--color-ink)] font-body placeholder:text-[var(--color-ink-soft)]/70 focus:outline-none";

export function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={cx(inputClass, props.className)} />;
}

export function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select {...props} className={cx(inputClass, "cursor-pointer appearance-none", props.className)} />
  );
}

export function EmptyState({
  icon,
  title,
  children,
}: {
  icon: ReactNode;
  title: string;
  children?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-12 text-center">
      <div
        className="clay-tint flex h-16 w-16 items-center justify-center rounded-xl text-[var(--color-grape-deep)]"
        style={vars({ "--tint": ACCENTS.grape })}
      >
        {icon}
      </div>
      <h3 className="text-lg">{title}</h3>
      {children && (
        <p className="max-w-sm text-sm text-[var(--color-ink-soft)]">{children}</p>
      )}
    </div>
  );
}

// ── Toasts ──────────────────────────────────────────────────────────────
interface Toast {
  id: number;
  message: string;
  tone: "ok" | "error";
}
const ToastContext = createContext<(message: string, tone?: "ok" | "error") => void>(() => {});
export const useToast = () => useContext(ToastContext);

export function ToastHost({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const push = useCallback((message: string, tone: "ok" | "error" = "ok") => {
    const id = Date.now() + Math.random();
    setToasts((t) => [...t, { id, message, tone }]);
    setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 3500);
  }, []);

  return (
    <ToastContext.Provider value={push}>
      {children}
      <div
        className="pointer-events-none fixed bottom-4 left-1/2 z-[1000] flex w-[min(92vw,26rem)] -translate-x-1/2 flex-col gap-2"
        aria-live="polite"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            className="animate-sheet pointer-events-auto flex items-center gap-2.5 rounded-lg border px-4 py-3 font-display text-sm font-semibold text-white shadow-sm"
            style={{
              background: t.tone === "ok" ? "var(--color-success)" : "var(--color-danger)",
              borderColor: t.tone === "ok" ? "#047857" : "#b91c1c",
            }}
            role="status"
          >
            {t.tone === "ok" ? <Check size={18} /> : <X size={18} />}
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
