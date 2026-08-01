import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Image } from "expo-image";
import { usePlaces } from "@/src/context/places-context";
import { getProxyImageUrl } from "@/services/api/urls";

export default function ResultsScreen() {
  const router = useRouter();
  const { id } = useLocalSearchParams<{ id?: string }>();
  const { getPlaceById } = usePlaces();

  const place = id ? getPlaceById(id) : undefined;

  const title = place?.title ?? "No result selected";
  const description = place?.summary ?? "There is no winning card yet.";

  const imageUrl = getProxyImageUrl(place?.photoName);

  return (
    <SafeAreaProvider style={styles.safeArea}>
      <View style={styles.container}>
        <Text style={styles.heading}>Winner</Text>

        <View style={styles.card}>
          <Text style={styles.title}>{title}</Text>
          <Text style={styles.description}>{description}</Text>

          <View style={styles.imageContainer}>
            <Image
              source={imageUrl ? { uri: imageUrl } : null}
              style={styles.image}
              contentFit="cover"
              transition={250}
            />
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
  imageContainer: {
    flex: 1,
    minHeight: 260,
    borderRadius: 18,
    borderWidth: 1,
    borderColor: "#d8d8d8",
    backgroundColor: "#f5f5f5",
    overflow: "hidden",
  },
  image: {
    width: "100%",
    height: "100%",
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
