import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <Stack>
      <Stack.Screen name="index" options={{ title: "Home" }} />
      <Stack.Screen name="room" options={{ title: "Room" }} />
      <Stack.Screen name="error" options={{ presentation: "modal", headerShown: false }} />
      <Stack.Screen name="join" options={{ title: "Join Room" }} />
      <Stack.Screen name="swipe" options={{ title: "Placeholder" }} />
    </Stack>
  );
}
