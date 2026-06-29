// Spacing, radii, and borders. Chunky 2px borders are the default look.

export const spacing = {
  0: 0,
  1: 4,
  2: 8,
  3: 12,
  4: 16,
  5: 20,
  6: 24,
  8: 32,
  10: 40,
  12: 48,
  16: 64,
  24: 96, // default section rhythm (`py-24` in the web spec)
} as const;

export const radius = {
  sm: 8,
  md: 16,
  lg: 24,
  full: 9999,
} as const;

// Speech-bubble blob: three rounded corners, one sharp.
export const blobRadius = {
  borderTopLeftRadius: radius.md,
  borderTopRightRadius: radius.md,
  borderBottomRightRadius: radius.md,
  borderBottomLeftRadius: 0,
} as const;

export const border = {
  width: 2, // chunky default
  hairline: 1,
} as const;

export const layout = {
  maxWidth: 1152, // max-w-6xl
} as const;

export type SpacingToken = keyof typeof spacing;
export type RadiusToken = keyof typeof radius;
