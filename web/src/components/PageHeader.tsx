import type { ReactNode } from "react";
import { ACCENTS, type Accent } from "./ui";

export function PageHeader({
  accent = "grape",
  icon,
  title,
  subtitle,
  action,
}: {
  accent?: Accent;
  icon: ReactNode;
  title: string;
  subtitle?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      <div className="flex items-center gap-3">
        <div
          className="clay-tint flex h-12 w-12 shrink-0 items-center justify-center rounded-lg"
          style={{
            ["--tint" as string]: ACCENTS[accent],
            color: `color-mix(in srgb, ${ACCENTS[accent]} 78%, black)`,
          }}
        >
          {icon}
        </div>
        <div className="min-w-0">
          <h1 className="truncate text-2xl sm:text-3xl">{title}</h1>
          {subtitle && (
            <p className="text-sm text-[var(--color-ink-soft)]">{subtitle}</p>
          )}
        </div>
      </div>
      {action}
    </div>
  );
}
