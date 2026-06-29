import { useEffect, useState } from "react";
import { ActivityIndicator, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useAuth } from "@/auth/AuthContext";
import { health } from "@/api/auth";
import { ApiError } from "@/api/client";
import { Button } from "@/components/Button";
import { Card } from "@/components/Card";
import {
  colors,
  fontFamily,
  fontSize,
  radius,
  border,
  spacing,
  layout,
  popShadow,
} from "@/theme";

export default function Home() {
  const { user, loading, signIn, signUp, signOut } = useAuth();
  const insets = useSafeAreaInsets();

  const [apiStatus, setApiStatus] = useState<string>("checking…");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [focused, setFocused] = useState<"email" | "password" | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    health()
      .then((r) => setApiStatus(r.status))
      .catch(() => setApiStatus("unreachable"));
  }, []);

  async function submit(action: (e: string, p: string) => Promise<void>) {
    setError(null);
    setBusy(true);
    try {
      await action(email.trim(), password);
      setPassword("");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Something went wrong");
    } finally {
      setBusy(false);
    }
  }

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.accent} />
      </View>
    );
  }

  const online = apiStatus === "ok";

  return (
    <ScrollView
      style={styles.screen}
      contentContainerStyle={[styles.container, { paddingBottom: insets.bottom + spacing[8] }]}
    >
      <View style={styles.inner}>
        {/* API status pill */}
        <View style={[styles.statusPill, popShadow(2)]}>
          <View style={[styles.dot, { backgroundColor: online ? colors.quaternary : colors.secondary }]} />
          <Text style={styles.statusText}>API · {apiStatus}</Text>
        </View>

        {user ? (
          <Card title="You're in 🎉" featured>
            <Text style={styles.body}>{user.email}</Text>
            <Text style={styles.meta}>id: {user.id}</Text>
            <Button label="Sign out" variant="secondary" onPress={signOut} />
          </Card>
        ) : (
          <Card title="Welcome">
            <Text style={styles.body}>Register or sign in to continue.</Text>

            <View style={styles.field}>
              <Text style={styles.label}>Email</Text>
              <TextInput
                style={[styles.input, focused === "email" && [styles.inputFocus, popShadow(4, colors.accent)]]}
                placeholder="you@example.com"
                placeholderTextColor={colors.mutedForeground}
                autoCapitalize="none"
                keyboardType="email-address"
                value={email}
                onChangeText={setEmail}
                onFocus={() => setFocused("email")}
                onBlur={() => setFocused(null)}
              />
            </View>

            <View style={styles.field}>
              <Text style={styles.label}>Password</Text>
              <TextInput
                style={[styles.input, focused === "password" && [styles.inputFocus, popShadow(4, colors.accent)]]}
                placeholder="min 8 characters"
                placeholderTextColor={colors.mutedForeground}
                secureTextEntry
                value={password}
                onChangeText={setPassword}
                onFocus={() => setFocused("password")}
                onBlur={() => setFocused(null)}
              />
            </View>

            {error ? <Text style={styles.error}>{error}</Text> : null}

            <Button label="Sign in" withIcon disabled={busy} onPress={() => submit(signIn)} />
            <Button label="Create account" variant="secondary" disabled={busy} onPress={() => submit(signUp)} />
          </Card>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { backgroundColor: colors.background },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.background },
  container: { padding: spacing[5], alignItems: "center" },
  inner: { width: "100%", maxWidth: layout.maxWidth / 2, gap: spacing[5] },
  statusPill: {
    alignSelf: "flex-start",
    flexDirection: "row",
    alignItems: "center",
    gap: spacing[2],
    paddingVertical: spacing[1],
    paddingHorizontal: spacing[3],
    borderRadius: radius.full,
    borderWidth: border.width,
    borderColor: colors.foreground,
    backgroundColor: colors.white,
  },
  dot: { width: 10, height: 10, borderRadius: radius.full },
  statusText: { fontFamily: fontFamily.bodyMedium, fontSize: fontSize.sm, color: colors.foreground },
  body: { fontFamily: fontFamily.body, fontSize: fontSize.base, color: colors.mutedForeground },
  meta: { fontFamily: fontFamily.body, fontSize: fontSize.xs, color: colors.mutedForeground },
  field: { gap: spacing[1] },
  label: {
    fontFamily: fontFamily.heading,
    fontSize: fontSize.xs,
    color: colors.foreground,
    textTransform: "uppercase",
    letterSpacing: 1,
  },
  input: {
    backgroundColor: colors.input,
    borderWidth: border.width,
    borderColor: colors.inputBorder,
    borderRadius: radius.sm,
    paddingHorizontal: spacing[4],
    paddingVertical: spacing[3],
    fontFamily: fontFamily.body,
    fontSize: fontSize.base,
    color: colors.foreground,
  },
  inputFocus: { borderColor: colors.accent },
  error: { fontFamily: fontFamily.bodyMedium, fontSize: fontSize.sm, color: colors.secondary },
});
