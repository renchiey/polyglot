import { useRef } from "react";
import {
  AccessibilityInfo,
  Animated,
  Pressable,
  StyleSheet,
  Text,
  View,
  type PressableProps,
} from "react-native";
import { ArrowRight } from "lucide-react-native";
import {
  colors,
  fontFamily,
  fontSize,
  radius,
  border,
  spacing,
  popShadow,
  duration,
} from "@/theme";

type Variant = "primary" | "secondary";

type Props = Omit<PressableProps, "children"> & {
  label: string;
  variant?: Variant;
  withIcon?: boolean;
};

// "Candy Button" — pill, chunky dark border, hard pop shadow that presses in.
// On press the element nudges down/right and the shadow shrinks (rest 4 -> 2),
// matching the design-system press behavior. Honors reduced-motion.
export function Button({ label, variant = "primary", withIcon = false, disabled, ...rest }: Props) {
  const press = useRef(new Animated.Value(0)).current;

  const animate = (to: number) =>
    AccessibilityInfo.isReduceMotionEnabled().then((reduced) => {
      if (reduced) return;
      Animated.timing(press, {
        toValue: to,
        duration: duration.fast,
        useNativeDriver: true,
      }).start();
    });

  const translate = press.interpolate({ inputRange: [0, 1], outputRange: [0, 2] });
  const isPrimary = variant === "primary";

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled: !!disabled }}
      disabled={disabled}
      onPressIn={() => animate(1)}
      onPressOut={() => animate(0)}
      {...rest}
    >
      <Animated.View
        style={[
          styles.base,
          isPrimary ? styles.primary : styles.secondary,
          popShadow(4),
          disabled && styles.disabled,
          { transform: [{ translateX: translate }, { translateY: translate }] },
        ]}
      >
        <Text style={[styles.label, isPrimary ? styles.primaryLabel : styles.secondaryLabel]}>
          {label}
        </Text>
        {withIcon && (
          <View style={styles.iconCircle}>
            <ArrowRight size={16} strokeWidth={2.5} color={colors.foreground} />
          </View>
        )}
      </Animated.View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: spacing[2],
    minHeight: 48,
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[6],
    borderRadius: radius.full,
    borderWidth: border.width,
    borderColor: colors.foreground,
  },
  primary: { backgroundColor: colors.accent },
  secondary: { backgroundColor: "transparent" },
  disabled: { opacity: 0.6 },
  label: { fontFamily: fontFamily.heading, fontSize: fontSize.base },
  primaryLabel: { color: colors.accentForeground },
  secondaryLabel: { color: colors.foreground },
  iconCircle: {
    width: 24,
    height: 24,
    borderRadius: radius.full,
    backgroundColor: colors.white,
    alignItems: "center",
    justifyContent: "center",
  },
});
