import { useState } from "react";
import { Brain, PartyPopper, Sparkles, Volume2 } from "lucide-react";
import { useApi } from "../lib/session";
import { useAsync } from "../lib/useAsync";
import { useSpeak } from "../lib/tts";
import type { DueCard, RecallResult, Rating } from "../lib/api";
import { Button, Card, EmptyState, Pill, Spinner, useToast, type Accent } from "../components/ui";
import { PageHeader } from "../components/PageHeader";

const RATINGS: { value: Rating; label: string; accent: Accent }[] = [
  { value: 1, label: "Again", accent: "danger" },
  { value: 2, label: "Hard", accent: "tangerine" },
  { value: 3, label: "Good", accent: "mint" },
  { value: 4, label: "Easy", accent: "sky" },
];


export function StudyView() {
  const api = useApi();
  const [mode, setMode] = useState<"review" | "quiz">("review");
  const due = useAsync(() => api.dueCards(50), []);

  return (
    <div className="space-y-5">
      <PageHeader
        accent="tangerine"
        icon={<Sparkles size={26} />}
        title="Study"
        subtitle="Spaced review and generative recall — exactly when you're about to forget."
      />

      <div className="flex gap-2.5">
        <Button accent="tangerine" soft={mode !== "review"} onClick={() => setMode("review")}>
          <Brain size={18} /> Review
        </Button>
        <Button accent="grape" soft={mode !== "quiz"} onClick={() => setMode("quiz")}>
          <Sparkles size={18} /> Recall quiz
        </Button>
      </div>

      {due.loading && <Card className="flex justify-center py-10"><Spinner /></Card>}
      {due.error && <Card><p className="text-[var(--color-danger)]">{due.error}</p></Card>}
      {due.data &&
        (mode === "review" ? (
          <ReviewMode cards={due.data} onChanged={due.reload} />
        ) : (
          <QuizMode cards={due.data} />
        ))}
    </div>
  );
}

const STATE_LABEL = ["New", "Learning", "Review", "Relearning"];

function ReviewMode({ cards, onChanged }: { cards: DueCard[]; onChanged: () => void }) {
  const api = useApi();
  const toast = useToast();
  const speak = useSpeak();
  const [idx, setIdx] = useState(0);
  const [revealed, setRevealed] = useState(false);
  const [busy, setBusy] = useState(false);

  if (cards.length === 0) {
    return (
      <Card>
        <EmptyState icon={<PartyPopper size={28} />} title="All caught up!">
          No cards are due right now. Add words in the Vault or read something new to grow your deck.
        </EmptyState>
      </Card>
    );
  }
  if (idx >= cards.length) {
    return (
      <Card>
        <EmptyState icon={<PartyPopper size={28} />} title={`Session complete — ${cards.length} reviewed`}>
          <Button accent="tangerine" onClick={onChanged}>Reload due cards</Button>
        </EmptyState>
      </Card>
    );
  }

  const card = cards[idx];

  async function rate(r: Rating) {
    setBusy(true);
    try {
      await api.review(card.word_id, r);
      setRevealed(false);
      setIdx((i) => i + 1);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Review failed", "error");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card className="space-y-5">
      <div className="flex items-center justify-between">
        <Pill accent="tangerine">{idx + 1} / {cards.length}</Pill>
        <Pill accent="grape">{STATE_LABEL[card.state] ?? "Card"}</Pill>
      </div>

      <div className="flex flex-col items-center gap-3 py-6 text-center">
        <button onClick={() => speak(card.term)} className="hanzi-chip font-hanzi text-7xl" aria-label={`Hear ${card.term}`}>
          {card.term}
        </button>
        {revealed ? (
          <Reveal term={card.term} translation={card.translation} />
        ) : (
          <Button accent="tangerine" soft onClick={() => setRevealed(true)}>Show answer</Button>
        )}
      </div>

      {revealed && (
        <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
          {RATINGS.map((r) => (
            <Button key={r.value} accent={r.accent} loading={busy} onClick={() => rate(r.value)}>
              {r.label}
            </Button>
          ))}
        </div>
      )}
    </Card>
  );
}

function Reveal({ term, translation }: { term: string; translation: string }) {
  const api = useApi();
  const { data, loading } = useAsync(() => api.lookup(term), [term]);
  return (
    <div className="animate-pop space-y-1">
      {loading ? (
        <Spinner />
      ) : (
        <>
          {data?.pinyin && <div className="font-display text-2xl text-[var(--color-grape-deep)]">{data.pinyin}</div>}
          <div className="text-lg text-[var(--color-ink)]">{translation || data?.definitions[0]}</div>
          {data && data.definitions.length > 1 && (
            <div className="text-sm text-[var(--color-ink-soft)]">{data.definitions.slice(1, 4).join(" · ")}</div>
          )}
        </>
      )}
    </div>
  );
}

function QuizMode({ cards }: { cards: DueCard[] }) {
  const api = useApi();
  const toast = useToast();
  const speak = useSpeak();
  const [i, setI] = useState(0);
  const [result, setResult] = useState<RecallResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [revealed, setRevealed] = useState(false);

  if (cards.length === 0) {
    return (
      <Card>
        <EmptyState icon={<Sparkles size={28} />} title="Nothing to quiz yet">
          The recall quiz builds a fresh sentence around a due word. Add and review words first.
        </EmptyState>
      </Card>
    );
  }

  async function nextQuiz() {
    const card = cards[i % cards.length];
    setLoading(true);
    setRevealed(false);
    setResult(null);
    try {
      setResult(await api.recall(card.word_id));
      setI((n) => n + 1);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Could not build quiz", "error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="space-y-5">
      <p className="text-sm text-[var(--color-ink-soft)]">
        A brand-new sentence built only from words you know — recall the missing one.
      </p>

      {result ? (
        <div className="animate-pop space-y-4">
          <div className="clay-inset p-5 text-center font-hanzi text-3xl leading-relaxed sm:text-4xl">
            {revealed ? result.sentence : result.cloze}
          </div>
          {revealed ? (
            <div className="text-center">
              <div className="font-hanzi text-5xl">{result.word}</div>
              {result.pinyin && <div className="font-display text-xl text-[var(--color-grape-deep)]">{result.pinyin}</div>}
            </div>
          ) : (
            <div className="flex justify-center gap-2.5">
              <Button accent="grape" soft onClick={() => speak(result.sentence)}>
                <Volume2 size={18} /> Hear it
              </Button>
              <Button accent="grape" onClick={() => setRevealed(true)}>Reveal answer</Button>
            </div>
          )}
        </div>
      ) : (
        <div className="py-6 text-center text-[var(--color-ink-soft)]">
          Tap below to generate your first recall sentence.
        </div>
      )}

      <Button accent="grape" onClick={nextQuiz} loading={loading}>
        <Sparkles size={18} /> {result ? "Next word" : "Start quiz"}
      </Button>
    </Card>
  );
}
