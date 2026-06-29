import { useCallback } from "react";
import { useSession } from "./session";
import { useSettings, type TtsSettings } from "./settings";

// browserSpeak is the fallback when the server has no Piper voice configured. It
// honours the user's speed/volume preferences directly on the utterance.
function browserSpeak(text: string, { speed, volume }: TtsSettings) {
  if (!("speechSynthesis" in window)) return;
  const u = new SpeechSynthesisUtterance(text);
  u.lang = "zh-CN";
  u.rate = 0.9 * speed;
  u.volume = volume;
  speechSynthesis.cancel();
  speechSynthesis.speak(u);
}

// useSpeak returns a speak(text) that plays server-side Piper audio, falling
// back to the browser's built-in Mandarin voice if Piper is unavailable. Speed
// is applied server-side (Piper preserves pitch); volume on the audio element.
export function useSpeak() {
  const { api } = useSession();
  const { tts } = useSettings();
  return useCallback(
    async (text: string) => {
      if (api) {
        try {
          const blob = await api.tts(text, tts.speed);
          const url = URL.createObjectURL(blob);
          const audio = new Audio(url);
          audio.volume = tts.volume;
          const cleanup = () => URL.revokeObjectURL(url);
          audio.onended = cleanup;
          audio.onerror = cleanup;
          await audio.play();
          return;
        } catch {
          /* Piper unavailable — fall through to browser TTS */
        }
      }
      browserSpeak(text, tts);
    },
    [api, tts],
  );
}
