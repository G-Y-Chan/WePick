import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { Card } from "./types";

export function SwipeCard({ card }: { card: Card }) {
  return (
    <View style={{ flex: 1 }}>
      <Text style={styles.title}>{card.title}</Text>
      <Text style={styles.description}>{card.description}</Text>

      <View style={styles.imagePlaceholder}>
        <Text style={styles.imagePlaceholderText}>Image Placeholder</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  title: { fontSize: 26, fontWeight: "800", marginBottom: 10 },
  description: { fontSize: 16, lineHeight: 22, marginBottom: 14 },
  imagePlaceholder: {
    flex: 1,
    borderRadius: 16,
    borderWidth: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  imagePlaceholderText: { fontSize: 14, opacity: 0.7 },
});
