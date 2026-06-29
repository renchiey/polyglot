// Bouncy, elastic feel. Values are tuned for react-native-reanimated / Animated,
// but kept library-agnostic so any animation layer can consume them.

export const duration = {
  fast: 150,
  base: 300,
  slow: 500,
} as const;

// Overshoot easing — equivalent to cubic-bezier(0.34, 1.56, 0.64, 1).
export const easing = {
  bounce: [0.34, 1.56, 0.64, 1] as const,
};

// Press feedback offsets for "Pop" elements (translate + shadow swap).
// Apply on the pressed state; pair with popShadowPresets.active.
export const pressTransform = {
  rest: { translateX: 0, translateY: 0 },
  active: { translateX: 2, translateY: 2 },
} as const;

// Wiggle keyframes (degrees) for icon hover/press: 0 -> 3 -> -3 -> 0.
export const wiggleKeyframes = [0, 3, -3, 0] as const;
