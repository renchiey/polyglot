import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";

// Token storage that uses SecureStore on native and localStorage on web,
// since expo-secure-store is unavailable in the browser.
export const tokenStorage = {
  async get(key: string): Promise<string | null> {
    if (Platform.OS === "web") {
      if (typeof localStorage === "undefined") return null;
      return localStorage.getItem(key);
    }
    return SecureStore.getItemAsync(key);
  },

  async set(key: string, value: string): Promise<void> {
    if (Platform.OS === "web") {
      localStorage.setItem(key, value);
      return;
    }
    await SecureStore.setItemAsync(key, value);
  },

  async remove(key: string): Promise<void> {
    if (Platform.OS === "web") {
      localStorage.removeItem(key);
      return;
    }
    await SecureStore.deleteItemAsync(key);
  },
};
