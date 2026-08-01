import React, { useEffect, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Button, StyleSheet, Text, View, ScrollView, TouchableOpacity, Switch } from "react-native";
import { router, useLocalSearchParams } from "expo-router";
import * as Location from "expo-location";

import { startRoom } from "../../services/api/http";
import { useRoom } from "@/src/context/room-context";

// Helper arrays for our filter UI
const RADIUS_OPTIONS = [
  { label: "1 mi", value: 1600 },
  { label: "5 mi", value: 8000 },
  { label: "15 mi", value: 24000 },
];

const PRICE_OPTIONS = [
  { label: "$", value: 1 },
  { label: "$$", value: 2 },
  { label: "$$$", value: 3 },
  { label: "$$$$", value: 4 },
];

const CATEGORY_OPTIONS = [
  { label: "Restaurant", value: "restaurant" },
  { label: "Cafe", value: "cafe" },
  { label: "Bar", value: "bar" },
];

export default function Room() {
  const { host } = useLocalSearchParams();
  const isHost = host === "true";

  const { status, lastMessage, roomCode } = useRoom();
  const [isStarting, setIsStarting] = useState(false);

  // Filter States
  const [radius, setRadius] = useState(8000); // Defaults to 5 miles
  const [maxPrice, setMaxPrice] = useState(2); // Defaults to $$
  const [category, setCategory] = useState("restaurant");
  const [openNow, setOpenNow] = useState(true);

  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.Header === "START") {
      router.replace("/swipe");
      return;
    }

    if (lastMessage.Header === "ERROR") {
      router.replace({
        pathname: "/error",
        params: {
          errorMessage: lastMessage.Body || "Server error",
        },
      });
    }
  }, [lastMessage]);

  const handleStartRoom = async () => {
    if (!roomCode) {
      router.replace({
        pathname: "/error",
        params: { errorMessage: "Missing room code" },
      });
      return;
    }

    if (status !== "connected") {
      router.replace({
        pathname: "/error",
        params: { errorMessage: "WebSocket is not connected yet" },
      });
      return;
    }

    try {
      setIsStarting(true);
      
      // 1. Request location permissions from the user
      const { status: permStatus } = await Location.requestForegroundPermissionsAsync();
      if (permStatus !== 'granted') {
        throw new Error("Location permission is required to find places near you.");
      }

      // 2. Fetch the current coordinates
      const location = await Location.getCurrentPositionAsync({});

      // 3. Bundle the filters with the retrieved coordinates
      const filters = {
        latitude: location.coords.latitude,
        longitude: location.coords.longitude,
        radius,
        maxPrice,
        category,
        openNow,
      };

      const response = await startRoom(roomCode, filters);
      const ok = response === "true";

      if (!ok) {
        router.replace({
          pathname: "/error",
          params: { errorMessage: "Unable to start room" },
        });
      }
    } catch (e: unknown) {
      let message = "Internal Server Error";
      if (e instanceof Error) {
        message = e.message;
      }
      router.replace({
        pathname: "/error",
        params: { errorMessage: message },
      });
    } finally {
      setIsStarting(false);
    }
  };

  const wsStatusText =
    status === "connecting"
      ? "Connecting..."
      : status === "connected"
      ? "Connected"
      : status === "error"
      ? "Connection error"
      : "";

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        {/* Fixed Header: Visible to everyone */}
        <View style={styles.header}>
          <Text style={styles.titleText}>Room Code: {roomCode}</Text>
          <Text style={styles.statusText}>{wsStatusText}</Text>
        </View>

        {isHost ? (
          <ScrollView style={styles.scrollArea} contentContainerStyle={styles.scrollContent}>
            <Text style={styles.sectionTitle}>Search Radius</Text>
            <View style={styles.row}>
              {RADIUS_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[styles.pill, radius === opt.value && styles.pillActive]}
                  onPress={() => setRadius(opt.value)}
                >
                  <Text style={[styles.pillText, radius === opt.value && styles.pillTextActive]}>
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <Text style={styles.sectionTitle}>Max Price</Text>
            <View style={styles.row}>
              {PRICE_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[styles.pill, maxPrice === opt.value && styles.pillActive]}
                  onPress={() => setMaxPrice(opt.value)}
                >
                  <Text style={[styles.pillText, maxPrice === opt.value && styles.pillTextActive]}>
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <Text style={styles.sectionTitle}>Category</Text>
            <View style={styles.row}>
              {CATEGORY_OPTIONS.map((opt) => (
                <TouchableOpacity
                  key={opt.value}
                  style={[styles.pill, category === opt.value && styles.pillActive]}
                  onPress={() => setCategory(opt.value)}
                >
                  <Text style={[styles.pillText, category === opt.value && styles.pillTextActive]}>
                    {opt.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <View style={[styles.row, styles.switchRow]}>
              <Text style={styles.sectionTitle}>Open Now</Text>
              <Switch value={openNow} onValueChange={setOpenNow} />
            </View>
          </ScrollView>
        ) : (
          <View style={styles.guestContainer}>
            <Text style={styles.guestText}>Waiting for host to set filters...</Text>
          </View>
        )}

        {/* Sticky Footer: Start Button */}
        {isHost && (
          <View style={styles.footer}>
            <Button
              title={isStarting ? "Fetching Location & Starting..." : "Start Session"}
              onPress={handleStartRoom}
              disabled={isStarting || status !== "connected"}
            />
          </View>
        )}
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
  },
  header: {
    paddingTop: 60,
    paddingBottom: 20,
    alignItems: "center",
    borderBottomWidth: 1,
    borderColor: "#eee",
  },
  titleText: {
    fontSize: 24,
    fontWeight: "bold",
  },
  statusText: {
    fontSize: 14,
    color: "#666",
    marginTop: 4,
  },
  scrollArea: {
    flex: 1,
  },
  scrollContent: {
    padding: 20,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: "600",
    marginBottom: 10,
    marginTop: 10,
  },
  row: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 10,
    marginBottom: 20,
  },
  switchRow: {
    justifyContent: "space-between",
    alignItems: "center",
    marginTop: 10,
  },
  pill: {
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 20,
    borderWidth: 1,
    borderColor: "#ccc",
    backgroundColor: "#f9f9f9",
  },
  pillActive: {
    backgroundColor: "#007AFF",
    borderColor: "#007AFF",
  },
  pillText: {
    fontSize: 14,
    color: "#333",
  },
  pillTextActive: {
    color: "#fff",
    fontWeight: "bold",
  },
  guestContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  guestText: {
    fontSize: 18,
    color: "#666",
  },
  footer: {
    padding: 20,
    borderTopWidth: 1,
    borderColor: "#eee",
    backgroundColor: "#fff",
  },
});