// Playful Geometric palette (light mode). See memory: design-system.
// `accent` drives primary actions; secondary/tertiary/quaternary rotate as
// decorative "confetti" (shapes, icons, emphasized words).
export const colors = {
  background: "#FFFDF5", // warm cream / paper
  foreground: "#1E293B", // slate 800 — also the hard-shadow color
  muted: "#F1F5F9", // slate 100
  mutedForeground: "#64748B", // slate 500

  accent: "#8B5CF6", // vivid violet (primary brand)
  accentForeground: "#FFFFFF",

  secondary: "#F472B6", // hot pink
  tertiary: "#FBBF24", // amber
  quaternary: "#34D399", // mint

  border: "#E2E8F0", // slate 200
  inputBorder: "#CBD5E1", // slate 300 (resting input border)
  input: "#FFFFFF",
  card: "#FFFFFF",
  ring: "#8B5CF6", // violet focus

  white: "#FFFFFF",
} as const;

// Rotate these for the "confetti" effect (decorative shapes/icons).
export const confetti = [colors.secondary, colors.tertiary, colors.quaternary, colors.accent] as const;

export type ColorToken = keyof typeof colors;
