// Playful Geometric theme. Import the aggregate `theme` for ergonomics, or pull
// individual token modules directly. See memory: design-system.
import { colors, confetti } from "./colors";
import { fontFamily, fontSize, lineHeight } from "./typography";
import { spacing, radius, blobRadius, border, layout } from "./layout";
import { duration, easing, pressTransform, wiggleKeyframes } from "./motion";

export const theme = {
  colors,
  confetti,
  fontFamily,
  fontSize,
  lineHeight,
  spacing,
  radius,
  blobRadius,
  border,
  layout,
  duration,
  easing,
  pressTransform,
  wiggleKeyframes,
} as const;

export type Theme = typeof theme;

export { colors, confetti } from "./colors";
export { fontFamily, fontSize, lineHeight } from "./typography";
export { spacing, radius, blobRadius, border, layout } from "./layout";
export { popShadow, cardShadow, popShadowPresets } from "./shadows";
export { duration, easing, pressTransform, wiggleKeyframes } from "./motion";
export { useAppFonts } from "./fonts";
