import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <Stack>
      <Stack.Screen name="index" options={{ title: "Home" }} />
      <Stack.Screen name="room" options={{ title: "Room" }} />
      <Stack.Screen name="join-room" options={{ title: "Join Room" }} />
    </Stack>
  );
}
