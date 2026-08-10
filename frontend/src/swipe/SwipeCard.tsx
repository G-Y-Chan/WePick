import React from "react";
import { View, Text, StyleSheet } from "react-native";
import { Image } from "expo-image";
import { Card } from "./types";
import { getImageUrl } from "@/services/api/urls";

export function SwipeCard({ card }: { card: Card }) {
  const imageUrl = getImageUrl(card.photoUrl, card.photoName);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>{card.title}</Text>
      
      {/* Category and Price Row */}
      <View style={styles.metaRow}>
        <Text style={styles.metaText}>
          {card.category}  •  {card.priceLevel}
        </Text>
        <Text style={[styles.openStatus, { color: card.openNow ? "#16a34a" : "#dc2626" }]}>
          {card.openNow ? "Open Now" : "Closed"}
        </Text>
      </View>

      {/* Rating and Reviews Row */}
      <View style={styles.ratingRow}>
        <Text style={styles.ratingText}>
          ⭐ {card.rating} ({card.reviewCount} reviews)
        </Text>
      </View>

      <Text style={styles.summary}>{card.summary}</Text>
      <Text style={styles.address}>📍 {card.address}</Text>

      <View style={styles.imageContainer}>
        <Image
          key={card.id}
          source={imageUrl ? { uri: imageUrl } : null}
          style={styles.image}
          contentFit="cover"
          transition={250}
          cachePolicy="memory-disk"
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { 
    flex: 1,
  },
  title: { 
    fontSize: 26, 
    fontWeight: "800", 
    marginBottom: 6,
  },
  metaRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 6,
  },
  metaText: {
    fontSize: 14,
    fontWeight: "600",
    color: "#555",
  },
  openStatus: {
    fontSize: 14,
    fontWeight: "bold",
  },
  ratingRow: {
    marginBottom: 12,
  },
  ratingText: {
    fontSize: 14,
    fontWeight: "500",
    color: "#333",
  },
  summary: { 
    fontSize: 16, 
    lineHeight: 22, 
    marginBottom: 10,
  },
  address: {
    fontSize: 14,
    color: "#666",
    fontStyle: "italic",
    marginBottom: 16,
  },
  imageContainer: {
    flex: 1,
    borderRadius: 16,
    overflow: "hidden",
    borderWidth: 1,
    borderColor: "#e5e5e5",
    backgroundColor: "#f5f5f5",
  },
  image: {
    width: "100%",
    height: "100%",
  },
});
