import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

// User-tunable playback preferences for text-to-speech. Speed is a multiplier of
// the voice's natural pace (sent to the server so Piper preserves pitch); volume
// is applied client-side to the audio element.
export interface TtsSettings {
  speed: number;
  volume: number;
}

export const TTS_DEFAULTS: TtsSettings = { speed: 1, volume: 1 };

// Allowed ranges, mirrored by the server's clamp so the UI can't request a pace
// the API will reject.
export const TTS_LIMITS = {
  speed: { min: 0.5, max: 2, step: 0.05 },
  volume: { min: 0, max: 1, step: 0.05 },
} as const;

const STORAGE_KEY = "context.tts-settings";

function clamp(n: number, min: number, max: number) {
  return Math.min(max, Math.max(min, n));
}

function load(): TtsSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return TTS_DEFAULTS;
    const parsed = JSON.parse(raw) as Partial<TtsSettings>;
    return {
      speed: clamp(Number(parsed.speed) || TTS_DEFAULTS.speed, TTS_LIMITS.speed.min, TTS_LIMITS.speed.max),
      volume: clamp(
        parsed.volume == null ? TTS_DEFAULTS.volume : Number(parsed.volume),
        TTS_LIMITS.volume.min,
        TTS_LIMITS.volume.max,
      ),
    };
  } catch {
    return TTS_DEFAULTS;
  }
}

interface SettingsValue {
  tts: TtsSettings;
  setTts: (patch: Partial<TtsSettings>) => void;
  resetTts: () => void;
}

const SettingsContext = createContext<SettingsValue | null>(null);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [tts, setState] = useState<TtsSettings>(load);

  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(tts));
    } catch {
      /* storage unavailable (private mode) — settings stay in-memory */
    }
  }, [tts]);

  const setTts = useCallback((patch: Partial<TtsSettings>) => {
    setState((prev) => ({ ...prev, ...patch }));
  }, []);

  const resetTts = useCallback(() => setState(TTS_DEFAULTS), []);

  const value = useMemo<SettingsValue>(() => ({ tts, setTts, resetTts }), [tts, setTts, resetTts]);

  return <SettingsContext.Provider value={value}>{children}</SettingsContext.Provider>;
}

export function useSettings(): SettingsValue {
  const ctx = useContext(SettingsContext);
  if (!ctx) throw new Error("useSettings must be used within SettingsProvider");
  return ctx;
}
