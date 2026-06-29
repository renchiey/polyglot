import { useState } from "react";
import { RotateCcw, Settings, Volume2, Volume1, VolumeX, Gauge, Play } from "lucide-react";
import { useSpeak } from "../lib/tts";
import { useSettings, TTS_DEFAULTS, TTS_LIMITS } from "../lib/settings";
import { Button, Card } from "../components/ui";
import { PageHeader } from "../components/PageHeader";

// A short Mandarin line used to preview the current voice settings.
const SAMPLE = "你好，欢迎来学习中文。";

export function SettingsView() {
  const { tts, setTts, resetTts } = useSettings();
  const speak = useSpeak();
  const [playing, setPlaying] = useState(false);

  async function preview() {
    setPlaying(true);
    try {
      await speak(SAMPLE);
    } finally {
      setPlaying(false);
    }
  }

  const VolIcon = tts.volume === 0 ? VolumeX : tts.volume < 0.5 ? Volume1 : Volume2;
  const isDefault = tts.speed === TTS_DEFAULTS.speed && tts.volume === TTS_DEFAULTS.volume;

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        accent="sky"
        icon={<Settings size={24} />}
        title="Settings"
        subtitle="Tune how spoken practice sounds."
      />

      <Card className="flex flex-col gap-7">
        <div className="flex items-center justify-between">
          <h2 className="text-lg">Voice playback</h2>
          <Button
            soft
            accent="sky"
            onClick={preview}
            loading={playing}
            className="flex items-center gap-2"
          >
            {!playing && <Play size={16} />}
            Play sample
          </Button>
        </div>

        <Slider
          icon={<Gauge size={20} />}
          accent="grape"
          label="Speed"
          value={tts.speed}
          {...TTS_LIMITS.speed}
          format={(v) => `${v.toFixed(2)}×`}
          onChange={(speed) => setTts({ speed })}
        />

        <Slider
          icon={<VolIcon size={20} />}
          accent="tangerine"
          label="Volume"
          value={tts.volume}
          {...TTS_LIMITS.volume}
          format={(v) => `${Math.round(v * 100)}%`}
          onChange={(volume) => setTts({ volume })}
        />

        <div className="flex items-center justify-between border-t border-[var(--color-ink)]/10 pt-4">
          <p className="text-sm text-[var(--color-ink-soft)]">
            Saved on this device. Speed adjusts pace without changing pitch.
          </p>
          <Button
            soft
            accent="sky"
            onClick={resetTts}
            disabled={isDefault}
            className="flex items-center gap-2"
          >
            <RotateCcw size={16} />
            Reset
          </Button>
        </div>
      </Card>
    </div>
  );
}

function Slider({
  icon,
  accent,
  label,
  value,
  min,
  max,
  step,
  format,
  onChange,
}: {
  icon: React.ReactNode;
  accent: string;
  label: string;
  value: number;
  min: number;
  max: number;
  step: number;
  format: (v: number) => string;
  onChange: (v: number) => void;
}) {
  return (
    <label className="block">
      <div className="mb-2 flex items-center justify-between">
        <span className="flex items-center gap-2 font-display font-bold text-[var(--color-ink)]">
          <span style={{ color: `var(--color-${accent})` }}>{icon}</span>
          {label}
        </span>
        <span className="clay-inset px-3 py-1 font-display text-sm font-bold tabular-nums text-[var(--color-ink)]">
          {format(value)}
        </span>
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full cursor-pointer accent-[var(--color-grape)]"
        style={{ accentColor: `var(--color-${accent})` }}
      />
    </label>
  );
}
