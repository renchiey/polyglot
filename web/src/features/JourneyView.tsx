import { useState } from "react";
import { ArrowRight, BookOpen, Check, Compass, Eye, EyeOff, GraduationCap, Mic, Play, Sparkles, Volume2, X } from "lucide-react";
import { useApi } from "../lib/session";
import { useAsync } from "../lib/useAsync";
import { useSpeak } from "../lib/tts";
import type { DictEntry, Journey, Rating } from "../lib/api";
import { Button, Card, Field, Pill, Spinner, TextInput, useToast, type Accent } from "../components/ui";
import { PageHeader } from "../components/PageHeader";

const RATINGS: { value: Rating; label: string; accent: Accent }[] = [
  { value: 1, label: "Again", accent: "danger" },
  { value: 2, label: "Hard", accent: "tangerine" },
  { value: 3, label: "Good", accent: "mint" },
  { value: 4, label: "Easy", accent: "sky" },
];
const PHASES = ["Read", "Learn", "Recall", "Talk"] as const;
type Phase = (typeof PHASES)[number];

type Learned = { term: string; wordId: string; pinyin?: string; definitions: string[]; characters?: DictEntry[] };

export function JourneyView() {
  const api = useApi();
  const toast = useToast();

  const [topic, setTopic] = useState("");
  const [starting, setStarting] = useState(false);
  const [journey, setJourney] = useState<Journey | null>(null);

  const [phase, setPhase] = useState<Phase>("Read");
  const [showPinyin, setShowPinyin] = useState(false);
  const [unknown, setUnknown] = useState<Set<string>>(new Set());
  const [learned, setLearned] = useState<Learned[]>([]);
  const [idx, setIdx] = useState(0);
  const [busy, setBusy] = useState(false);

  async function start() {
    setStarting(true);
    try {
      const j = await api.startJourney(topic.trim());
      setJourney(j);
      setPhase("Read");
      setShowPinyin(false);
      setUnknown(new Set());
      setLearned([]);
      setIdx(0);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Could not start journey", "error");
    } finally {
      setStarting(false);
    }
  }

  function toggleUnknown(term: string) {
    setUnknown((prev) => {
      const next = new Set(prev);
      next.has(term) ? next.delete(term) : next.add(term);
      return next;
    });
  }

  // Read → Learn: marked unknowns plus a few topic-related suggestions, each
  // looked up and added to the vault.
  async function goLearn() {
    if (!journey) return;
    const terms = [...new Set([...unknown, ...journey.suggestions.map((s) => s.term)])];
    if (terms.length === 0) {
      setPhase("Talk");
      return;
    }
    setBusy(true);
    try {
      const results = (
        await Promise.all(
          terms.map(async (term): Promise<Learned | null> => {
            try {
              const lk = await api.lookup(term);
              const word = await api.createWord(term, lk.definitions[0] ?? "", lk.definitions.join("; "));
              return { term, wordId: word.id, pinyin: lk.pinyin, definitions: lk.definitions, characters: lk.characters };
            } catch {
              return null;
            }
          }),
        )
      ).filter((x): x is Learned => x !== null);

      if (results.length === 0) {
        toast("Couldn't look those up — skipping ahead.", "error");
        setPhase("Talk");
        return;
      }
      setLearned(results);
      setIdx(0);
      setPhase("Learn");
      toast(`${results.length} word${results.length > 1 ? "s" : ""} added to your Vault`);
    } finally {
      setBusy(false);
    }
  }

  async function rateRecall(rating: Rating) {
    const word = learned[idx];
    try {
      await api.review(word.wordId, rating);
    } catch {
      /* best-effort scheduling */
    }
    if (idx + 1 < learned.length) setIdx(idx + 1);
    else setPhase("Talk");
  }

  if (!journey) {
    return (
      <div className="space-y-5">
        <Header />
        <Card className="space-y-4">
          <Field label="Topic (optional)" hint="Leave blank for an everyday-life story at your level.">
            <TextInput value={topic} onChange={(e) => setTopic(e.target.value)} placeholder="ordering coffee" />
          </Field>
          <Button accent="bubble" onClick={start} loading={starting}>
            <Play size={18} /> Begin today's journey
          </Button>
        </Card>
      </div>
    );
  }

  const activeIdx = PHASES.indexOf(phase);

  return (
    <div className="space-y-5">
      <Header />
      <div className="flex gap-2">
        {PHASES.map((p, i) => (
          <div
            key={p}
            className={[
              "flex-1 rounded-lg border px-2 py-2 text-center font-display text-sm font-semibold",
              i === activeIdx ? "clay-tint" : "border-[var(--color-line)] text-[var(--color-ink-soft)]",
            ].join(" ")}
            style={i === activeIdx ? { ["--tint" as string]: "var(--color-bubble)" } : undefined}
          >
            {i + 1}. {p}
          </div>
        ))}
      </div>

      {phase === "Read" && (
        <ReadPhase
          journey={journey}
          showPinyin={showPinyin}
          onTogglePinyin={() => setShowPinyin((v) => !v)}
          unknown={unknown}
          onToggleWord={toggleUnknown}
          busy={busy}
          onLearn={goLearn}
        />
      )}

      {phase === "Learn" && learned[idx] && (
        <LearnPhase
          word={learned[idx]}
          index={idx}
          total={learned.length}
          onNext={() => (idx + 1 < learned.length ? setIdx(idx + 1) : (setIdx(0), setPhase("Recall")))}
        />
      )}

      {phase === "Recall" && learned[idx] && (
        <RecallPhase key={learned[idx].wordId} word={learned[idx]} index={idx} total={learned.length} onRate={rateRecall} />
      )}

      {phase === "Talk" && <TalkPhase journey={journey} learnedCount={learned.length} onRestart={() => setJourney(null)} />}
    </div>
  );
}

function Header() {
  return (
    <PageHeader
      accent="bubble"
      icon={<Compass size={26} />}
      title="Daily Journey"
      subtitle="Read, mark what you don't know, learn it, then use it."
    />
  );
}

function ReadPhase({
  journey,
  showPinyin,
  onTogglePinyin,
  unknown,
  onToggleWord,
  busy,
  onLearn,
}: {
  journey: Journey;
  showPinyin: boolean;
  onTogglePinyin: () => void;
  unknown: Set<string>;
  onToggleWord: (term: string) => void;
  busy: boolean;
  onLearn: () => void;
}) {
  const speak = useSpeak();
  const learnCount = new Set([...unknown, ...journey.suggestions.map((s) => s.term)]).size;

  return (
    <>
      <Card className="space-y-3">
        <div className="flex items-center justify-between">
          <Pill accent="sky">Story · HSK {journey.level}</Pill>
          <div className="flex gap-1.5">
            <button onClick={onTogglePinyin} className="clay-btn clay-btn-soft !rounded-lg !px-2.5 !py-2 !text-xs" aria-pressed={showPinyin}>
              {showPinyin ? <Eye size={16} /> : <EyeOff size={16} />} Pinyin
            </button>
            <button onClick={() => speak(journey.story.map((t) => t.text).join(""))} className="clay-btn clay-btn-soft !rounded-lg !p-2" aria-label="Hear the story">
              <Volume2 size={18} />
            </button>
          </div>
        </div>
        <p className="flex flex-wrap items-end gap-x-1 gap-y-3 font-hanzi text-2xl leading-relaxed sm:text-3xl">
          {journey.story.map((tok, i) =>
            tok.cjk ? (
              <button
                key={i}
                onClick={() => onToggleWord(tok.text)}
                className={["hanzi-chip", unknown.has(tok.text) ? "!border-[var(--color-bubble)] !bg-[color-mix(in_srgb,var(--color-bubble)_12%,white)]" : ""].join(" ")}
                aria-pressed={unknown.has(tok.text)}
              >
                {showPinyin && (
                  <span className="font-display text-[0.6rem] font-semibold leading-none text-[var(--color-grape-deep)]">{tok.pinyin || "·"}</span>
                )}
                <span>{tok.text}</span>
              </button>
            ) : (
              <span key={i} className="self-end whitespace-pre-wrap text-[var(--color-ink-soft)]">
                {tok.text}
              </span>
            ),
          )}
        </p>
        <p className="text-xs text-[var(--color-ink-soft)]">Tap any word you don't know to mark it — you'll learn those, plus a few new ones, next.</p>
      </Card>

      <Card className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-[var(--color-ink-soft)]">
          {unknown.size > 0
            ? `${unknown.size} marked${journey.suggestions.length ? ` + ${journey.suggestions.length} new to discover` : ""}.`
            : journey.suggestions.length
              ? `${journey.suggestions.length} new word${journey.suggestions.length > 1 ? "s" : ""} to discover.`
              : "Mark the words you don't know, or skip ahead."}
        </p>
        <Button accent="bubble" onClick={onLearn} loading={busy}>
          {learnCount > 0 ? (
            <>
              <GraduationCap size={18} /> Learn {learnCount}
            </>
          ) : (
            <>I knew them all</>
          )}
        </Button>
      </Card>
    </>
  );
}

function LearnPhase({ word, index, total, onNext }: { word: Learned; index: number; total: number; onNext: () => void }) {
  const speak = useSpeak();
  return (
    <Card className="animate-pop space-y-4">
      <Pill accent="sky">
        <BookOpen size={14} /> Learn {index + 1} of {total}
      </Pill>
      <div className="flex items-center gap-3">
        <span className="font-hanzi text-5xl text-[var(--color-ink)]">{word.term}</span>
        {word.pinyin && <span className="font-display text-2xl text-[var(--color-grape-deep)]">{word.pinyin}</span>}
        <button onClick={() => speak(word.term)} className="clay-btn clay-btn-soft !rounded-lg !p-2" aria-label="Hear the word">
          <Volume2 size={18} />
        </button>
      </div>
      <ul className="space-y-1.5">
        {word.definitions.slice(0, 5).map((d, i) => (
          <li key={i} className="flex gap-2 text-[15px] text-[var(--color-ink)]">
            <span className="font-display text-[var(--color-ink-soft)]">{i + 1}.</span>
            {d}
          </li>
        ))}
      </ul>
      {word.characters && word.characters.length > 0 && (
        <div>
          <p className="mb-2 font-display text-sm font-semibold text-[var(--color-ink-soft)]">Character breakdown</p>
          <div className="flex flex-wrap gap-2">
            {word.characters.map((c) => (
              <div key={c.word} className="clay-tint flex items-center gap-2 rounded-lg px-3 py-2" style={{ ["--tint" as string]: "var(--color-sky)" }}>
                <span className="font-hanzi text-2xl">{c.word}</span>
                <span>
                  <span className="block font-display text-sm text-[var(--color-grape-deep)]">{c.pinyin}</span>
                  <span className="block max-w-[10rem] truncate text-xs text-[var(--color-ink-soft)]">{c.definitions[0]}</span>
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
      <Button accent="bubble" onClick={onNext}>
        {index + 1 < total ? (
          <>
            Next word <ArrowRight size={18} />
          </>
        ) : (
          <>
            Practise using them <ArrowRight size={18} />
          </>
        )}
      </Button>
    </Card>
  );
}

function RecallPhase({ word, index, total, onRate }: { word: Learned; index: number; total: number; onRate: (r: Rating) => void }) {
  const api = useApi();
  const speak = useSpeak();
  const { data, loading, error } = useAsync(() => api.recall(word.wordId), [word.wordId]);
  const [revealed, setRevealed] = useState(false);
  const [guess, setGuess] = useState("");
  const [correct, setCorrect] = useState<boolean | null>(null);

  function check() {
    if (!data) return;
    setCorrect(guess.trim() === data.word);
    setRevealed(true);
  }

  return (
    <Card className="animate-pop space-y-4">
      <Pill accent="bubble">
        <Sparkles size={14} /> Use it {index + 1} of {total}
      </Pill>
      <div>
        <p className="font-display text-sm font-semibold text-[var(--color-ink-soft)]">Recall the word that means…</p>
        <p className="text-2xl font-semibold text-[var(--color-ink)]">{word.definitions[0] || word.term}</p>
      </div>

      {loading && <div className="flex justify-center py-6"><Spinner /></div>}
      {error && <p className="text-sm text-[var(--color-danger)]">{error}</p>}

      {data && (
        <>
          <div className="clay-inset p-4 font-hanzi text-2xl leading-relaxed text-[var(--color-ink)]">
            {revealed ? data.sentence : data.cloze}
          </div>

          {!revealed ? (
            <div className="space-y-2.5">
              <div className="flex gap-2">
                <TextInput
                  value={guess}
                  onChange={(e) => setGuess(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && check()}
                  placeholder="Type the word (in Chinese)…"
                  className="font-hanzi text-lg"
                />
                <Button accent="bubble" onClick={check}>Check</Button>
              </div>
              <button onClick={() => setRevealed(true)} className="text-sm font-semibold text-[var(--color-ink-soft)] underline">
                or reveal the answer
              </button>
            </div>
          ) : (
            <div className="animate-pop space-y-3">
              {correct !== null && (
                <div
                  className="flex items-center gap-2 font-display font-bold"
                  style={{ color: correct ? "var(--color-success)" : "var(--color-danger)" }}
                >
                  {correct ? <Check size={18} /> : <X size={18} />}
                  {correct ? "Correct!" : `You typed “${guess || "—"}”`}
                </div>
              )}
              <div className="flex items-center gap-3">
                <span className="font-hanzi text-4xl text-[var(--color-ink)]">{data.word}</span>
                {data.pinyin && <span className="font-display text-xl text-[var(--color-grape-deep)]">{data.pinyin}</span>}
                <button onClick={() => speak(data.sentence)} className="clay-btn clay-btn-soft !rounded-lg !p-2" aria-label="Hear the sentence">
                  <Volume2 size={18} />
                </button>
              </div>
              <p className="text-sm text-[var(--color-ink-soft)]">How well did you recall it?</p>
              <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
                {RATINGS.map((rt) => (
                  <Button key={rt.value} accent={rt.accent} onClick={() => onRate(rt.value)}>
                    {rt.label}
                  </Button>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </Card>
  );
}

function TalkPhase({ journey, learnedCount, onRestart }: { journey: Journey; learnedCount: number; onRestart: () => void }) {
  return (
    <Card className="animate-pop space-y-3">
      <Pill accent="bubble">
        <Mic size={14} /> Interaction
      </Pill>
      <h3 className="text-lg">
        {learnedCount > 0 ? `Nice — ${learnedCount} new word${learnedCount > 1 ? "s" : ""} in your Vault.` : "Story complete."}
      </h3>
      <p className="text-[var(--color-ink-soft)]">{journey.interaction.opening}</p>
      <div className="flex flex-wrap gap-2.5">
        <Button accent="sky" disabled>
          <Mic size={18} /> Start voice chat (soon)
        </Button>
        <Button accent="bubble" soft onClick={onRestart}>
          New journey
        </Button>
      </div>
    </Card>
  );
}
