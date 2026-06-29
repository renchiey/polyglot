import { useFonts } from "expo-font";
import { Outfit_700Bold, Outfit_800ExtraBold } from "@expo-google-fonts/outfit";
import {
  PlusJakartaSans_400Regular,
  PlusJakartaSans_500Medium,
} from "@expo-google-fonts/plus-jakarta-sans";

// Loads the design-system fonts. The keys here MUST match fontFamily in
// typography.ts so the registered family names line up with what styles use.
export function useAppFonts() {
  const [loaded, error] = useFonts({
    Outfit_700Bold,
    Outfit_800ExtraBold,
    PlusJakartaSans_400Regular,
    PlusJakartaSans_500Medium,
  });
  return { fontsLoaded: loaded, fontError: error };
}
