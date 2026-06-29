// Font family keys map to the names registered by useAppFonts() (see fonts.ts).
// Headings: Outfit (geometric, friendly). Body: Plus Jakarta Sans (humanist).
export const fontFamily = {
  heading: "Outfit_700Bold",
  headingExtra: "Outfit_800ExtraBold",
  body: "PlusJakartaSans_400Regular",
  bodyMedium: "PlusJakartaSans_500Medium",
} as const;

// Type scale — Major Third (1.25) from a 16px base.
export const fontSize = {
  xs: 13,
  sm: 14,
  base: 16,
  lg: 20,
  xl: 25,
  "2xl": 31,
  "3xl": 39,
  "4xl": 49,
} as const;

export const lineHeight = {
  xs: 18,
  sm: 20,
  base: 24,
  lg: 28,
  xl: 34,
  "2xl": 40,
  "3xl": 48,
  "4xl": 58,
} as const;

export type FontSizeToken = keyof typeof fontSize;
