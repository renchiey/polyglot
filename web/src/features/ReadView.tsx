import { useState } from "react";
import { BookOpen, Check, Eye, EyeOff, Mic, Volume2, Wand2 } from "lucide-react";
import { useApi } from "../lib/session";
import { useSpeak } from "../lib/tts";
import type { AssessResult, SegToken } from "../lib/api";
import { Button, Card, Spinner, useToast } from "../components/ui";
import { PageHeader } from "../components/PageHeader";

// The learner picks a group of topics; each generation draws one at random, and
// the text style is chosen at random too, so they meet varied input.
const TOPICS = [
  "daily life",
  "food & drink",
  "travel",
  "family & friends",
  "work & study",
  "shopping",
  "weather & seasons",
  "hobbies",
  "health",
  "city life",
];
const KINDS = ["short story", "dialogue", "news snippet", "diary entry", "a description"];

const pick = <T,>(arr: T[]) => arr[Math.floor(Math.random() * arr.length)];

export function ReadView() {
  const api = useApi();
  const toast = useToast();

  const [topics, setTopics] = useState<Set<string>>(new Set(["daily life", "food & drink", "travel"]));
  const [showPinyin, setShowPinyin] = useState(false);
  const [loading, setLoading] = useState(false);
  const [tokens, setTokens] = useState<SegToken[] | null>(null);
  const [marked, setMarked] = useState<Set<string>>(new Set());

  function toggleTopic(t: string) {
    setTopics((prev) => {
      const next = new Set(prev);
      next.has(t) ? next.delete(t) : next.add(t);
      return next;
    });
  }

  async function generate() {
    const chosen = topics.size > 0 ? [...topics] : TOPICS;
    setLoading(true);
    setTokens(null);
    setMarked(new Set());
    try {
      const gen = await api.generate({ language: "zh", target_level: 0, topic: pick(chosen), kind: pick(KINDS) });
      const seg = await api.segment(gen.text);
      setTokens(seg.tokens);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Generation failed", "error");
    } finally {
      setLoading(false);
    }
  }

  // Clicking an unknown word highlights it and saves it to the vault for review.
  async function markWord(term: string) {
    if (marked.has(term)) return;
    setMarked((prev) => new Set(prev).add(term));
    try {
      const lk = await api.lookup(term).catch(() => null);
      await api.createWord(term, lk?.definitions[0] ?? "", lk?.definitions.join("; ") ?? "");
      toast(`${term} saved to your Vault`);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Could not save", "error");
    }
  }

  const sentences = tokens ? splitSentences(tokens) : [];
  const fullText = tokens ? tokens.map((t) => t.text).join("") : "";

  return (
    <div className="space-y-5">
      <PageHeader
        accent="sky"
        icon={<BookOpen size={26} />}
        title="Reading Room"
        subtitle="A fresh text at your level. Read without pinyin; tap words you don't know."
      />

      <Card className="space-y-4">
        <div>
          <p className="mb-2 font-display text-sm font-bold text-[var(--color-ink)]">Topics</p>
          <div className="flex flex-wrap gap-2">
            {TOPICS.map((t) => {
              const on = topics.has(t);
              return (
                <button
                  key={t}
                  onClick={() => toggleTopic(t)}
                  className={[
                    "rounded-lg border px-3 py-1.5 text-sm font-semibold transition-colors",
                    on ? "clay-tint" : "border-[var(--color-line)] text-[var(--color-ink-soft)] hover:border-[var(--color-line-strong)]",
                  ].join(" ")}
                  style={on ? { ["--tint" as string]: "var(--color-sky)" } : undefined}
                  aria-pressed={on}
                >
                  {t}
                </button>
              );
            })}
          </div>
          <p className="mt-2 text-xs text-[var(--color-ink-soft)]">A random topic and style are chosen each time, at your current level.</p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button accent="sky" onClick={generate} loading={loading}>
            <Wand2 size={18} /> {tokens ? "New text" : "Generate"}
          </Button>
          <button onClick={() => setShowPinyin((v) => !v)} className="clay-btn clay-btn-soft" aria-pressed={showPinyin}>
            {showPinyin ? <Eye size={18} /> : <EyeOff size={18} />} Pinyin {showPinyin ? "on" : "off"}
          </button>
        </div>
      </Card>

      {loading && <Card className="flex justify-center py-10"><Spinner /></Card>}

      {tokens && (
        <>
          <Card className="space-y-4">
            <p className="text-xs text-[var(--color-ink-soft)]">Tap a word you don't know to save it. Use the mic to read each line aloud.</p>
            <div className="space-y-3">
              {sentences.map((sent, i) => (
                <SentenceBlock key={i} tokens={sent} showPinyin={showPinyin} marked={marked} onWord={markWord} />
              ))}
            </div>
          </Card>

          <TranslationBox text={fullText} />
        </>
      )}
    </div>
  );
}

function SentenceBlock({
  tokens,
  showPinyin,
  marked,
  onWord,
}: {
  tokens: SegToken[];
  showPinyin: boolean;
  marked: Set<string>;
  onWord: (term: string) => void;
}) {
  const speak = useSpeak();
  const sentence = tokens.map((t) => t.text).join("");
  const [score, setScore] = useState<number | null>(null);
  const [listening, setListening] = useState(false);

  async function readAloud() {
    setScore(null);
    setListening(true);
    try {
      const heard = await recognizeMandarin();
      setScore(similarity(sentence, heard));
    } catch (e) {
      const msg = e instanceof Error ? e.message : "mic error";
      setScore(msg === "unsupported" ? -1 : 0);
    } finally {
      setListening(false);
    }
  }

  return (
    <div className="flex items-start gap-2">
      <div className="flex flex-col gap-1 pt-1">
        <button onClick={() => speak(sentence)} className="clay-btn clay-btn-soft !rounded-lg !p-1.5" aria-label="Hear line">
          <Volume2 size={16} />
        </button>
        <button
          onClick={readAloud}
          className="clay-btn !rounded-lg !p-1.5"
          style={{ ["--btn" as string]: listening ? "var(--color-danger)" : "var(--color-sky)" }}
          aria-label="Read aloud"
        >
          <Mic size={16} />
        </button>
      </div>
      <div className="min-w-0">
        <p className="flex flex-wrap items-end gap-x-1 gap-y-2 font-hanzi text-2xl leading-relaxed">
          {tokens.map((tok, i) =>
            tok.cjk ? (
              <button
                key={i}
                onClick={() => onWord(tok.text)}
                className={["hanzi-chip", marked.has(tok.text) ? "!border-[var(--color-sky)] !bg-[color-mix(in_srgb,var(--color-sky)_12%,white)]" : ""].join(" ")}
                aria-pressed={marked.has(tok.text)}
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
        {listening && <p className="mt-1 text-xs font-semibold text-[var(--color-danger)]">Listening… read the line aloud.</p>}
        {score !== null && !listening && (
          <p
            className="mt-1 text-xs font-semibold"
            style={{ color: score < 0 ? "var(--color-ink-soft)" : score >= 70 ? "var(--color-success)" : "var(--color-tangerine)" }}
          >
            {score < 0 ? "Read-aloud needs Chrome." : `Pronunciation match: ${score}%`}
          </p>
        )}
      </div>
    </div>
  );
}

function TranslationBox({ text }: { text: string }) {
  const api = useApi();
  const toast = useToast();
  const [value, setValue] = useState("");
  const [assessing, setAssessing] = useState(false);
  const [result, setResult] = useState<AssessResult | null>(null);

  async function assess() {
    if (!value.trim()) return;
    setAssessing(true);
    setResult(null);
    try {
      setResult(await api.assessTranslation(text, value.trim()));
    } catch (e) {
      toast(e instanceof Error ? e.message : "Assessment failed", "error");
    } finally {
      setAssessing(false);
    }
  }

  return (
    <Card className="space-y-3">
      <div>
        <h3 className="text-lg">Translate it</h3>
        <p className="text-sm text-[var(--color-ink-soft)]">Write what the text means in English — we'll gauge your comprehension.</p>
      </div>
      <textarea
        value={value}
        onChange={(e) => setValue(e.target.value)}
        rows={3}
        placeholder="Your English translation…"
        className="clay-inset w-full resize-y px-4 py-2.5 font-body text-[var(--color-ink)] placeholder:text-[var(--color-ink-soft)]/70 focus:outline-none"
      />
      <Button accent="sky" onClick={assess} loading={assessing} disabled={!value.trim()}>
        <Check size={18} /> Check my understanding
      </Button>
      {result && (
        <div className="clay-tint animate-pop rounded-lg p-4" style={{ ["--tint" as string]: result.score >= 70 ? "var(--color-success)" : "var(--color-tangerine)" }}>
          <div className="font-display text-2xl font-bold text-[var(--color-ink)]">{result.score}/100</div>
          <p className="text-sm text-[var(--color-ink)]">{result.feedback}</p>
        </div>
      )}
    </Card>
  );
}

// splitSentences groups tokens into sentences on CJK sentence-ending punctuation.
function splitSentences(tokens: SegToken[]): SegToken[][] {
  const out: SegToken[][] = [];
  let cur: SegToken[] = [];
  for (const t of tokens) {
    if (t.text.trim() === "") continue;
    cur.push(t);
    if (/[。！？!?]/.test(t.text)) {
      out.push(cur);
      cur = [];
    }
  }
  if (cur.length) out.push(cur);
  return out;
}

// recognizeMandarin captures one utterance via the Web Speech API (Chrome).
// SpeechRecognition isn't in TS's DOM lib (vendor-prefixed), so it's untyped.
function recognizeMandarin(): Promise<string> {
  /* eslint-disable @typescript-eslint/no-explicit-any */
  const w = window as any;
  const SR = w.webkitSpeechRecognition || w.SpeechRecognition;
  if (!SR) return Promise.reject(new Error("unsupported"));
  return new Promise<string>((resolve, reject) => {
    const rec = new SR();
    rec.lang = "zh-CN";
    rec.interimResults = false;
    rec.maxAlternatives = 1;
    rec.onresult = (e: any) => resolve(e.results[0][0].transcript as string);
    rec.onerror = (e: any) => reject(new Error(e.error || "mic"));
    rec.start();
  });
}

// similarity is a char-level Dice coefficient over CJK characters, as a percent.
function similarity(a: string, b: string): number {
  const ca = [...a].filter(isHan);
  const cb = [...b].filter(isHan);
  if (ca.length === 0) return 0;
  const pool = new Map<string, number>();
  for (const c of cb) pool.set(c, (pool.get(c) ?? 0) + 1);
  let common = 0;
  for (const c of ca) {
    const n = pool.get(c) ?? 0;
    if (n > 0) {
      common++;
      pool.set(c, n - 1);
    }
  }
  return Math.round((200 * common) / (ca.length + cb.length));
}

const isHan = (c: string) => c >= "一" && c <= "鿿";
