import { Platform } from "react-native";
import { colors } from "./colors";

// Minimal shape so the helper is assignable to both ViewStyle and TextStyle
// arrays (RN's ViewStyle/TextStyle disagree on some props, but both accept
// boxShadow as a string).
type ShadowStyle = { boxShadow: string };

// The signature "Pop" shadow: hard, offset, zero-blur. RN 0.76+ (new arch) and
// react-native-web both support the CSS `boxShadow` style string, which is the
// only way to get a spread/zero-blur offset shadow cross-platform.
//
// On mobile we halve the offset per the responsive spec ("reduce pop shadows to
// 2px"). Pass `force` to keep the full offset on small screens (e.g. web).
export function popShadow(offset = 4, color: string = colors.foreground): ShadowStyle {
  const o = Platform.OS === "web" ? offset : Math.max(2, Math.round(offset / 2));
  return { boxShadow: `${o}px ${o}px 0px ${color}` };
}

// Soft hard-shadow for sticker cards (light slate, or pink when featured).
export function cardShadow(color: string = colors.border): ShadowStyle {
  return popShadow(8, color);
}

// Convenience presets mirroring the design-system button states.
export const popShadowPresets = {
  rest: popShadow(4),
  hover: popShadow(6),
  active: popShadow(2),
} as const;
