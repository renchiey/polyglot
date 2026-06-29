import { useState } from "react";
import { Boxes, Clock, Pencil, Plus, Search, Trash2, X } from "lucide-react";
import { useApi } from "../lib/session";
import { useAsync } from "../lib/useAsync";
import type { Word } from "../lib/api";
import { Button, Card, EmptyState, Field, Spinner, TextInput, useToast } from "../components/ui";
import { PageHeader } from "../components/PageHeader";
import { LookupSheet } from "../components/LookupSheet";

export function VaultView() {
  const api = useApi();
  const toast = useToast();
  const { data, loading, error, setData, reload } = useAsync(() => api.listWords(), []);
  const [lookup, setLookup] = useState<string | null>(null);
  const [editing, setEditing] = useState<Word | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Word | null>(null);

  const [term, setTerm] = useState("");
  const [translation, setTranslation] = useState("");
  const [adding, setAdding] = useState(false);

  async function add(e: React.FormEvent) {
    e.preventDefault();
    if (!term.trim()) return;
    setAdding(true);
    try {
      const word = await api.createWord(term.trim(), translation.trim());
      // Re-adding an existing word returns it (card reset) — move it to the top
      // rather than duplicating the row.
      setData([word, ...(data ?? []).filter((x) => x.id !== word.id)]);
      setTerm("");
      setTranslation("");
      toast(`${word.term} added`);
    } catch (err) {
      toast(err instanceof Error ? err.message : "Could not add", "error");
    } finally {
      setAdding(false);
    }
  }

  async function doDelete(w: Word) {
    setConfirmDelete(null);
    try {
      await api.deleteWord(w.id);
      setData((data ?? []).filter((x) => x.id !== w.id));
      toast(`${w.term} removed`);
    } catch (err) {
      toast(err instanceof Error ? err.message : "Could not delete", "error");
    }
  }

  return (
    <div className="space-y-5">
      <PageHeader
        accent="mint"
        icon={<Boxes size={26} />}
        title="The Vault"
        subtitle="Your growing word collection. Mastered words steer what you read next."
        action={data ? <span className="pill" style={{ ["--tint" as string]: "var(--color-mint)" }}>{data.length} words</span> : undefined}
      />

      <Card>
        <form onSubmit={add} className="grid items-end gap-3 sm:grid-cols-[1fr_1fr_auto]">
          <Field label="Word (Hanzi)">
            <TextInput value={term} onChange={(e) => setTerm(e.target.value)} placeholder="咖啡" required />
          </Field>
          <Field label="Translation">
            <TextInput value={translation} onChange={(e) => setTranslation(e.target.value)} placeholder="coffee" />
          </Field>
          <Button accent="mint" type="submit" loading={adding}>
            <Plus size={18} /> Add
          </Button>
        </form>
      </Card>

      {loading && <Card className="flex justify-center py-10"><Spinner /></Card>}
      {error && <Card><p className="text-[var(--color-danger)]">{error}</p></Card>}

      {data && data.length === 0 && (
        <Card>
          <EmptyState icon={<Boxes size={28} />} title="Your Vault is empty">
            Add a word above, or tap words while reading to collect them here.
          </EmptyState>
        </Card>
      )}

      {data && data.length > 0 && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((w) => (
            <Card key={w.id} className="animate-pop flex flex-col gap-3 !p-4">
              <div className="flex items-start justify-between gap-2">
                <button onClick={() => setLookup(w.term)} className="text-left">
                  <div className="font-hanzi text-4xl leading-tight text-[var(--color-ink)]">{w.term}</div>
                  {w.translation && <div className="mt-1 text-[var(--color-ink-soft)]">{w.translation}</div>}
                </button>
                <div className="flex gap-1.5">
                  <IconBtn label="Look up" onClick={() => setLookup(w.term)}><Search size={16} /></IconBtn>
                  <IconBtn label="Edit" onClick={() => setEditing(w)}><Pencil size={16} /></IconBtn>
                  <IconBtn label="Delete" danger onClick={() => setConfirmDelete(w)}><Trash2 size={16} /></IconBtn>
                </div>
              </div>
              <NextReview iso={w.next_review} />
            </Card>
          ))}
        </div>
      )}

      {confirmDelete && (
        <div className="fixed inset-0 z-[900] flex items-center justify-center bg-[#1c1b22]/50 p-4 backdrop-blur-sm" onClick={() => setConfirmDelete(null)}>
          <div className="clay animate-pop w-full max-w-sm space-y-4 p-6 text-center" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg">Remove this word?</h3>
            <div className="font-hanzi text-4xl text-[var(--color-ink)]">{confirmDelete.term}</div>
            <p className="text-sm text-[var(--color-ink-soft)]">
              It will be removed from your Vault and review schedule.
            </p>
            <div className="flex gap-2.5">
              <Button accent="grape" soft className="flex-1" onClick={() => setConfirmDelete(null)}>
                Cancel
              </Button>
              <Button accent="danger" className="flex-1" onClick={() => doDelete(confirmDelete)}>
                <Trash2 size={18} /> Remove
              </Button>
            </div>
          </div>
        </div>
      )}

      {lookup && <LookupSheet word={lookup} onClose={() => setLookup(null)} />}
      {editing && (
        <EditModal
          word={editing}
          onClose={() => setEditing(null)}
          onSaved={(updated) => {
            setData((data ?? []).map((x) => (x.id === updated.id ? updated : x)));
            setEditing(null);
            void reload;
          }}
        />
      )}
    </div>
  );
}

// NextReview shows when a word is next due for review (its FSRS card's due
// time), relative to now. Words due now are flagged to nudge a study session.
function NextReview({ iso }: { iso?: string }) {
  const due = iso ? new Date(iso).getTime() - Date.now() : 0;
  const overdue = due <= 0;
  return (
    <div
      className="mt-auto flex items-center gap-1.5 border-t border-[var(--color-line)] pt-2 text-xs font-semibold"
      style={{ color: overdue ? "var(--color-tangerine)" : "var(--color-ink-soft)" }}
    >
      <Clock size={13} />
      {overdue ? "Due for review" : `Next review ${relative(due)}`}
    </div>
  );
}

function relative(ms: number): string {
  const mins = Math.round(ms / 60000);
  if (mins < 60) return `in ${mins}m`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `in ${hrs}h`;
  const days = Math.round(hrs / 24);
  if (days < 30) return `in ${days}d`;
  return `in ${Math.round(days / 30)}mo`;
}

function IconBtn({
  label,
  onClick,
  danger,
  children,
}: {
  label: string;
  onClick: () => void;
  danger?: boolean;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      title={label}
      className="clay-btn clay-btn-soft !rounded-xl !p-2"
      style={danger ? { color: "var(--color-danger)" } : undefined}
    >
      {children}
    </button>
  );
}

function EditModal({
  word,
  onClose,
  onSaved,
}: {
  word: Word;
  onClose: () => void;
  onSaved: (w: Word) => void;
}) {
  const api = useApi();
  const toast = useToast();
  const [term, setTerm] = useState(word.term);
  const [translation, setTranslation] = useState(word.translation);
  const [definition, setDefinition] = useState(word.definition);
  const [saving, setSaving] = useState(false);

  async function save() {
    setSaving(true);
    try {
      const updated = await api.updateWord(word.id, term.trim(), translation, definition);
      onSaved(updated);
      toast("Saved");
    } catch (e) {
      toast(e instanceof Error ? e.message : "Could not save", "error");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 z-[900] flex items-center justify-center bg-[#1c1b22]/50 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="clay animate-pop w-full max-w-sm space-y-4 p-6" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg">Edit word</h3>
          <button onClick={onClose} aria-label="Close" className="clay-btn clay-btn-soft !rounded-lg !p-2"><X size={16} /></button>
        </div>
        <Field label="Word"><TextInput value={term} onChange={(e) => setTerm(e.target.value)} /></Field>
        <Field label="Translation"><TextInput value={translation} onChange={(e) => setTranslation(e.target.value)} /></Field>
        <Field label="Notes"><TextInput value={definition} onChange={(e) => setDefinition(e.target.value)} /></Field>
        <Button accent="mint" onClick={save} loading={saving}>Save changes</Button>
      </div>
    </div>
  );
}
