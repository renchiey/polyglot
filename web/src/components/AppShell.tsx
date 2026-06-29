import { type ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { BookOpen, Boxes, Compass, Home, Languages, Settings, Sparkles } from "lucide-react";
import { ACCENTS, type Accent } from "./ui";

const NAV: { to: string; label: string; icon: ReactNode; accent: Accent }[] = [
  { to: "/", label: "Home", icon: <Home size={22} />, accent: "grape" },
  { to: "/journey", label: "Journey", icon: <Compass size={22} />, accent: "bubble" },
  { to: "/read", label: "Read", icon: <BookOpen size={22} />, accent: "sky" },
  { to: "/study", label: "Study", icon: <Sparkles size={22} />, accent: "tangerine" },
  { to: "/vault", label: "Vault", icon: <Boxes size={22} />, accent: "mint" },
  { to: "/settings", label: "Settings", icon: <Settings size={22} />, accent: "sky" },
];

function navClass(active: boolean) {
  return [
    "flex items-center gap-3 rounded-lg px-3.5 py-2.5 font-display font-bold transition-transform",
    active ? "clay-tint" : "text-[var(--color-ink-soft)] hover:text-[var(--color-ink)]",
  ].join(" ");
}

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-7xl gap-6 px-4 pb-24 pt-4 md:px-6 md:pb-6">
      {/* Desktop sidebar */}
      <aside className="sticky top-4 hidden h-[calc(100dvh-2rem)] w-60 shrink-0 flex-col md:flex">
        <Brand />
        <nav className="clay mt-4 flex flex-1 flex-col gap-1.5 p-3">
          {NAV.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              end={n.to === "/"}
              style={{ ["--tint" as string]: ACCENTS[n.accent] }}
              className={({ isActive }) => navClass(isActive)}
            >
              {n.icon}
              {n.label}
            </NavLink>
          ))}
          <div className="mt-auto px-2 pt-3 text-xs leading-relaxed text-[var(--color-ink-soft)]">
            Boss Fights & Voice are on the roadmap.
          </div>
        </nav>
      </aside>

      <div className="min-w-0 flex-1">
        <header className="mb-5 flex items-center md:hidden">
          <Brand />
        </header>
        {children}
      </div>

      {/* Mobile bottom nav */}
      <nav className="clay fixed inset-x-3 bottom-3 z-50 flex items-center justify-around rounded-xl px-2 py-2 md:hidden">
        {NAV.map((n) => (
          <NavLink
            key={n.to}
            to={n.to}
            end={n.to === "/"}
            style={{ ["--tint" as string]: ACCENTS[n.accent] }}
            className={({ isActive }) =>
              [
                "flex min-w-14 flex-col items-center gap-0.5 rounded-lg px-2 py-1.5 text-[0.7rem] font-bold transition-transform",
                isActive ? "clay-tint" : "text-[var(--color-ink-soft)]",
              ].join(" ")
            }
          >
            {n.icon}
            {n.label}
          </NavLink>
        ))}
      </nav>
    </div>
  );
}

function Brand() {
  return (
    <div className="flex items-center gap-2.5">
      <div
        className="flex h-11 w-11 items-center justify-center rounded-lg text-white"
        style={{ background: "var(--color-grape)", border: "1px solid var(--color-grape-deep)" }}
      >
        <Languages size={22} />
      </div>
      <div className="leading-tight">
        <div className="font-display text-xl font-bold text-[var(--color-ink)]">Polyglot</div>
        <div className="text-[0.7rem] font-bold text-[var(--color-ink-soft)]">learn in context</div>
      </div>
    </div>
  );
}
