// Typed client for the Polyglot Go API. One createApi() instance is bound to a
// base URL + bearer token by the session provider; views call its methods.

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// ── Response shapes (mirror server/internal/handlers) ───────────────────
export interface AuditReport {
  passed: boolean;
  language: string;
  target_level: number;
  sentence_level: number;
  out_of_bounds: Record<string, number>;
  unknown: string[];
}
export interface GenerateResult {
  text: string;
  passed: boolean;
  rounds: number;
  report: AuditReport;
}
export interface SegToken {
  text: string;
  pinyin?: string;
  cjk: boolean;
}
export interface DictEntry {
  word: string;
  pinyin: string;
  definitions: string[];
}
export interface LookupResult extends DictEntry {
  characters?: DictEntry[];
}
export interface Word {
  id: string;
  term: string;
  translation: string;
  definition: string;
  created_at: string;
  updated_at: string;
  next_review?: string;
}
export interface DueCard {
  word_id: string;
  term: string;
  translation: string;
  due: string;
  state: number;
  reps: number;
  lapses: number;
  scheduled_days: number;
}
export interface CardView {
  word_id: string;
  term: string;
  translation: string;
  due: string;
  state: number;
  reps: number;
  lapses: number;
  scheduled_days: number;
}
export interface RecallResult {
  word: string;
  pinyin?: string;
  sentence: string;
  cloze: string;
  passed: boolean;
  rounds: number;
}
export interface AssessResult {
  score: number;
  feedback: string;
}
export interface AuthResponse {
  token: string;
  user: { id: string; email: string; createdAt: string };
}
export interface Progress {
  vocabulary: number;
  syntax: number;
  listening: number;
  speaking: number;
  recommended_level: number;
}
export interface JourneyToken {
  text: string;
  pinyin?: string;
  cjk: boolean;
}
export interface JourneyInteraction {
  persona: string;
  opening: string;
  voice: boolean;
}
export interface JourneySuggestion {
  term: string;
  pinyin?: string;
  gloss?: string;
}
export interface Journey {
  id: string;
  topic: string;
  level: number;
  passed: boolean;
  story: JourneyToken[];
  suggestions: JourneySuggestion[];
  interaction: JourneyInteraction;
}

export type GenerateBody = {
  language: string;
  target_level: number;
  topic?: string;
  kind?: string;
};
export type Rating = 1 | 2 | 3 | 4;

async function request<T>(
  base: string,
  token: string | null,
  path: string,
  opts: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(opts.headers as Record<string, string>),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  let res: Response;
  try {
    res = await fetch(base + path, { ...opts, headers });
  } catch {
    throw new ApiError("Can't reach the server. Is the API running?", 0);
  }
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(message, res.status);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

const q = (params: Record<string, string>) =>
  "?" + new URLSearchParams(params).toString();

/** Bootstrap helpers (no token needed). */
export const auth = {
  health: (base: string) => request<{ status: string }>(base, null, "/health"),
  login: (base: string, email: string, password: string) =>
    request<AuthResponse>(base, null, "/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  register: (base: string, email: string, password: string) =>
    request<AuthResponse>(base, null, "/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
};

export type Api = ReturnType<typeof createApi>;

export function createApi(base: string, token: string) {
  const r = <T>(path: string, opts?: RequestInit) =>
    request<T>(base, token, path, opts);

  return {
    base,
    generate: (body: GenerateBody) =>
      r<GenerateResult>("/generate", { method: "POST", body: JSON.stringify(body) }),
    segment: (text: string, language = "zh") =>
      r<{ tokens: SegToken[] }>("/segment", {
        method: "POST",
        body: JSON.stringify({ language, text }),
      }),
    lookup: (word: string, language = "zh") =>
      r<LookupResult>("/lookup" + q({ word, language })),
    listWords: () => r<Word[]>("/words"),
    createWord: (term: string, translation = "", definition = "") =>
      r<Word>("/words", {
        method: "POST",
        body: JSON.stringify({ term, translation, definition }),
      }),
    updateWord: (id: string, term: string, translation: string, definition: string) =>
      r<Word>(`/words/${id}`, {
        method: "PUT",
        body: JSON.stringify({ term, translation, definition }),
      }),
    deleteWord: (id: string) => r<void>(`/words/${id}`, { method: "DELETE" }),
    dueCards: (limit = 50) => r<DueCard[]>("/cards/due" + q({ limit: String(limit) })),
    review: (wordId: string, rating: Rating) =>
      r<CardView>("/cards/review", {
        method: "POST",
        body: JSON.stringify({ word_id: wordId, rating }),
      }),
    recall: (wordId: string) =>
      r<RecallResult>("/recall", {
        method: "POST",
        body: JSON.stringify({ word_id: wordId }),
      }),
    progress: () => r<Progress>("/progress"),
    startJourney: (topic = "", targetLevel = 0) =>
      r<Journey>("/journey/start", {
        method: "POST",
        body: JSON.stringify({ topic, target_level: targetLevel }),
      }),
    assessTranslation: (text: string, translation: string) =>
      r<AssessResult>("/assess/translation", {
        method: "POST",
        body: JSON.stringify({ text, translation }),
      }),
    tts: async (text: string, speed = 1): Promise<Blob> => {
      const res = await fetch(base + "/tts", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ text, speed }),
      });
      if (!res.ok) throw new ApiError("tts failed", res.status);
      return res.blob();
    },
  };
}
