import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { useLocalSearchParams, useRouter } from "expo-router";

export default function ResultsScreen() {
  const router = useRouter();
  const params = useLocalSearchParams<{
    Title?: string;
    Description?: string;
  }>();

  const title = params.Title ?? "No result selected";
  const description = params.Description ?? "There is no winning card yet.";

  return (
    <SafeAreaProvider style={styles.safeArea}>
      <View style={styles.container}>
        <Text style={styles.heading}>Winner</Text>

        <View style={styles.card}>
          <Text style={styles.title}>{title}</Text>
          <Text style={styles.description}>{description}</Text>

          <View style={styles.imagePlaceholder}>
            <Text style={styles.imagePlaceholderText}>Winning Card Image</Text>
          </View>
        </View>

        <Pressable style={styles.button} onPress={() => router.back()}>
          <Text style={styles.buttonText}>Back</Text>
        </Pressable>
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: "#fff",
  },
  container: {
    flex: 1,
    paddingHorizontal: 20,
    paddingTop: 24,
    paddingBottom: 20,
  },
  kicker: {
    fontSize: 14,
    fontWeight: "600",
    color: "#666",
    marginBottom: 6,
    textTransform: "uppercase",
    letterSpacing: 1,
  },
  heading: {
    fontSize: 32,
    fontWeight: "800",
    marginBottom: 20,
    color: "#111",
    textAlign: "center",
  },
  card: {
    flex: 1,
    backgroundColor: "#f8f8f8",
    borderRadius: 24,
    padding: 20,
    borderWidth: 1,
    borderColor: "#e5e5e5",
  },
  badge: {
    alignSelf: "flex-start",
    fontSize: 13,
    fontWeight: "700",
    color: "#444",
    backgroundColor: "#eaeaea",
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 999,
    marginBottom: 14,
  },
  title: {
    fontSize: 28,
    fontWeight: "800",
    marginBottom: 12,
    color: "#111",
  },
  description: {
    fontSize: 16,
    lineHeight: 24,
    color: "#444",
    marginBottom: 18,
  },
  imagePlaceholder: {
    flex: 1,
    minHeight: 260,
    borderRadius: 18,
    borderWidth: 1,
    borderColor: "#d8d8d8",
    backgroundColor: "#fff",
    justifyContent: "center",
    alignItems: "center",
  },
  imagePlaceholderText: {
    fontSize: 14,
    color: "#777",
  },
  button: {
    marginTop: 18,
    backgroundColor: "#111",
    paddingVertical: 16,
    borderRadius: 14,
    alignItems: "center",
  },
  buttonText: {
    color: "#fff",
    fontSize: 16,
    fontWeight: "700",
  },
});