import type { ReactNode } from "react";
import { StyleSheet, Text, View, type ViewStyle } from "react-native";
import {
  colors,
  fontFamily,
  fontSize,
  radius,
  border,
  spacing,
  cardShadow,
} from "@/theme";

type Props = {
  title?: string;
  featured?: boolean; // pink shadow instead of slate
  style?: ViewStyle;
  children?: ReactNode;
};

// "Sticker" card — white surface, chunky dark border, soft hard shadow
// (slate by default, hot-pink when featured).
export function Card({ title, featured = false, style, children }: Props) {
  return (
    <View style={[styles.card, cardShadow(featured ? colors.secondary : colors.border), style]}>
      {title ? <Text style={styles.title}>{title}</Text> : null}
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.card,
    borderWidth: border.width,
    borderColor: colors.foreground,
    borderRadius: radius.lg,
    padding: spacing[5],
    gap: spacing[3],
  },
  title: {
    fontFamily: fontFamily.headingExtra,
    fontSize: fontSize.xl,
    color: colors.foreground,
  },
});
