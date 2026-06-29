import { useEffect, useState } from "react";
import { BookmarkPlus, Volume2, X } from "lucide-react";
import { useApi } from "../lib/session";
import { useAsync } from "../lib/useAsync";
import { useSpeak } from "../lib/tts";
import { Button, Spinner, useToast } from "./ui";

export function LookupSheet({ word, onClose }: { word: string; onClose: () => void }) {
  const [current, setCurrent] = useState(word);
  const api = useApi();
  const toast = useToast();
  const speak = useSpeak();

  const { data, loading, error } = useAsync(() => api.lookup(current), [current]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const addToVault = async () => {
    if (!data) return;
    try {
      await api.createWord(current, data.definitions[0] ?? "", data.definitions.join("; "));
      toast(`${current} added to your Vault`);
    } catch (e) {
      toast(e instanceof Error ? e.message : "Could not add", "error");
    }
  };

  return (
    <div
      className="fixed inset-0 z-[900] flex items-end justify-center bg-[#1c1b22]/50 p-0 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label={`Dictionary entry for ${current}`}
    >
      <div
        className="clay animate-sheet max-h-[88vh] w-full overflow-y-auto rounded-b-none rounded-t-xl p-6 sm:max-w-md sm:rounded-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between">
          <div>
            <div className="font-hanzi text-5xl leading-none text-[var(--color-ink)]">{current}</div>
            {data && <div className="mt-2 font-display text-xl text-[var(--color-grape-deep)]">{data.pinyin}</div>}
          </div>
          <button onClick={onClose} aria-label="Close" className="clay-btn clay-btn-soft !rounded-lg !p-2.5">
            <X size={18} />
          </button>
        </div>

        {loading && <div className="flex justify-center py-8"><Spinner /></div>}
        {error && <p className="py-4 text-sm text-[var(--color-danger)]">{error}</p>}

        {data && (
          <>
            <ul className="mb-4 space-y-1.5">
              {data.definitions.map((d, i) => (
                <li key={i} className="flex gap-2 text-[15px] text-[var(--color-ink)]">
                  <span className="font-display text-[var(--color-ink-soft)]">{i + 1}.</span>
                  {d}
                </li>
              ))}
            </ul>

            {data.characters && data.characters.length > 0 && (
              <div className="mb-5">
                <p className="mb-2 font-display text-sm font-bold text-[var(--color-ink-soft)]">Character breakdown</p>
                <div className="flex flex-wrap gap-2">
                  {data.characters.map((c) => (
                    <button
                      key={c.word}
                      onClick={() => setCurrent(c.word)}
                      className="clay-tint flex items-center gap-2 rounded-lg px-3 py-2 text-left"
                      style={{ ["--tint" as string]: "var(--color-sky)" }}
                    >
                      <span className="font-hanzi text-2xl">{c.word}</span>
                      <span>
                        <span className="block font-display text-sm text-[var(--color-grape-deep)]">{c.pinyin}</span>
                        <span className="block max-w-[10rem] truncate text-xs text-[var(--color-ink-soft)]">
                          {c.definitions[0]}
                        </span>
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="flex flex-wrap gap-2.5">
              <Button accent="sky" onClick={() => speak(current)}>
                <Volume2 size={18} /> Hear it
              </Button>
              <Button accent="mint" onClick={addToVault}>
                <BookmarkPlus size={18} /> Add to Vault
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
