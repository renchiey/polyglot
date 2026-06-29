import { Link, useNavigate } from "react-router-dom";
import {
  BookOpen,
  Boxes,
  Gauge,
  Lock,
  Mic,
  Play,
  Sparkles,
  Trophy,
} from "lucide-react";
import { useApi } from "../lib/session";
import { useAsync } from "../lib/useAsync";
import { ACCENTS, Button, Card, Pill, type Accent } from "../components/ui";

export function DashboardView() {
  const api = useApi();
  const navigate = useNavigate();
  const words = useAsync(() => api.listWords(), []);
  const due = useAsync(() => api.dueCards(100), []);

  return (
    <div className="space-y-5">
      {/* Hero — the Daily Journey entry point */}
      <div className="clay border-l-[5px] p-7 sm:p-9" style={{ borderLeftColor: "var(--color-grape)" }}>
        <div className="max-w-2xl">
          <Pill accent="grape">Daily Journey</Pill>
          <h1 className="mt-3 text-3xl leading-[1.12] sm:text-[2.6rem]">
            Read it, notice it, recall it — one workflow.
          </h1>
          <p className="mt-3 max-w-xl text-[var(--color-ink-soft)]">
            Reading and spaced recall tuned to exactly what you know, built on the science of
            comprehensible input.
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Button accent="grape" onClick={() => navigate("/journey")}>
              <Play size={18} /> Start today's journey
            </Button>
            <Button accent="grape" soft onClick={() => navigate("/study")}>
              <Sparkles size={18} /> Review words
            </Button>
          </div>
        </div>
      </div>

      <ProgressCard />

      {/* Live stats */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <StatTile accent="mint" icon={<Boxes size={22} />} label="Words in Vault" value={words.data?.length} to="/vault" />
        <StatTile accent="tangerine" icon={<Sparkles size={22} />} label="Due to review" value={due.data?.length} to="/study" />
        <StatTile accent="sky" icon={<BookOpen size={22} />} label="Reading Room" value="Open" to="/read" />
      </div>

      {/* Roadmap — extendable surfaces, not yet wired */}
      <div>
        <h2 className="mb-2.5 px-1 text-lg">Coming soon</h2>
        <div className="grid gap-3 sm:grid-cols-2">
          <SoonTile accent="bubble" icon={<Trophy size={20} />} title="Boss Fights" note="Real-world task challenges" />
          <SoonTile accent="sky" icon={<Mic size={20} />} title="Voice Chat" note="Talk with story characters" />
        </div>
      </div>
    </div>
  );
}

const VECTORS = [
  ["Vocabulary", "vocabulary", "grape"],
  ["Syntax", "syntax", "sky"],
  ["Listening", "listening", "mint"],
  ["Speaking", "speaking", "tangerine"],
] as const;

// ProgressCard surfaces Linguistic Elo. Only Vocabulary moves today; the others
// are baseline until their signals (grammar, voice) come online.
function ProgressCard() {
  const api = useApi();
  const { data } = useAsync(() => api.progress(), []);
  if (!data) return null;
  return (
    <Card className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-lg">
          <Gauge size={20} /> Your level
        </h2>
        <Pill accent="grape">HSK {data.recommended_level} recommended</Pill>
      </div>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {VECTORS.map(([label, key, accent]) => (
          <div
            key={key}
            className="clay-tint rounded-lg p-3"
            style={{ ["--tint" as string]: ACCENTS[accent] }}
          >
            <div className="font-mono text-2xl font-semibold text-[var(--color-ink)]">{data[key]}</div>
            <div className="text-xs font-semibold text-[var(--color-ink-soft)]">{label}</div>
          </div>
        ))}
      </div>
    </Card>
  );
}

function StatTile({
  accent,
  icon,
  label,
  value,
  to,
}: {
  accent: Accent;
  icon: React.ReactNode;
  label: string;
  value: number | string | undefined;
  to: string;
}) {
  return (
    <Link to={to}>
      <Card className="flex h-full items-center gap-3 !p-4 transition-colors hover:border-[var(--color-line-strong)]">
        <div
          className="clay-tint flex h-12 w-12 items-center justify-center rounded-lg"
          style={{ ["--tint" as string]: ACCENTS[accent], color: `color-mix(in srgb, ${ACCENTS[accent]} 78%, black)` }}
        >
          {icon}
        </div>
        <div>
          <div className="font-display text-2xl font-bold text-[var(--color-ink)]">
            {value ?? "—"}
          </div>
          <div className="text-sm text-[var(--color-ink-soft)]">{label}</div>
        </div>
      </Card>
    </Link>
  );
}

function SoonTile({
  accent,
  icon,
  title,
  note,
}: {
  accent: Accent;
  icon: React.ReactNode;
  title: string;
  note: string;
}) {
  return (
    <div className="clay flex items-start gap-3 p-4 opacity-75">
      <div
        className="clay-tint flex h-11 w-11 shrink-0 items-center justify-center rounded-lg"
        style={{ ["--tint" as string]: ACCENTS[accent], color: `color-mix(in srgb, ${ACCENTS[accent]} 78%, black)` }}
      >
        {icon}
      </div>
      <div>
        <div className="flex items-center gap-1.5 font-display font-bold text-[var(--color-ink)]">
          {title} <Lock size={13} className="text-[var(--color-ink-soft)]" />
        </div>
        <div className="text-xs text-[var(--color-ink-soft)]">{note}</div>
      </div>
    </div>
  );
}
