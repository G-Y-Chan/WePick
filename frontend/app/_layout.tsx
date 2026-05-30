import { PlacesProvider } from "@/src/context/places-context";
import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <PlacesProvider>
      <Stack>
        <Stack.Screen name="index" options={{ title: "Home" }} />
        <Stack.Screen name="join" options={{ title: "Join Room" }} />
        <Stack.Screen name="error" options={{ presentation: "modal", headerShown: false }} />
        <Stack.Screen name="(room)" options={{ headerShown: false }} />
      </Stack>
    </PlacesProvider>
  );
}