import React, { useMemo, useRef, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  Animated,
  PanResponder,
  useWindowDimensions,
} from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";

type Card = {
  id: string;
  title: string;
  description: string;
};

const SWIPE_THRESHOLD_RATIO = 0.25;
const SWIPE_OUT_DURATION_MS = 180;

export default function Swipe() {
  const { width, height } = useWindowDimensions();

  const data = useMemo<Card[]>(
    () =>
      Array.from({ length: 10 }, (_, i) => ({
        id: String(i + 1),
        title: `Place ${i + 1}`,
        description:
          "Placeholder description. Swipe right to accept, left to reject.",
      })),
    []
  );

  const [index, setIndex] = useState(0);
  const position = useRef(new Animated.ValueXY({ x: 0, y: 0 })).current;

  const swipeThreshold = width * SWIPE_THRESHOLD_RATIO;

  // Rotate top card slightly
  const rotate = position.x.interpolate({
    inputRange: [-width, 0, width],
    outputRange: ["-12deg", "0deg", "12deg"],
  });

  // Edge glow opacity (ramps up as you drag)
  const leftGlowOpacity = position.x.interpolate({
    inputRange: [-swipeThreshold * 1.8, -swipeThreshold, 0],
    outputRange: [0.35, 1, 0],
    extrapolate: "clamp",
  });

  const rightGlowOpacity = position.x.interpolate({
    inputRange: [0, swipeThreshold, swipeThreshold * 1.8],
    outputRange: [0, 0.35, 1],
    extrapolate: "clamp",
  });

  // Next card scales slightly for stacking effect
  const nextScale = position.x.interpolate({
    inputRange: [-width, 0, width],
    outputRange: [0.98, 0.95, 0.98],
    extrapolate: "clamp",
  });

  function resetPosition() {
    Animated.spring(position, {
      toValue: { x: 0, y: 0 },
      useNativeDriver: false,
      friction: 6,
    }).start();
  }

  function forceSwipe(direction: "left" | "right") {
    const x = direction === "right" ? width * 1.2 : -width * 1.2;

    Animated.timing(position, {
      toValue: { x, y: 0 },
      duration: SWIPE_OUT_DURATION_MS,
      useNativeDriver: false,
    }).start(() => onSwipeComplete(direction));
  }

  function onSwipeComplete(direction: "left" | "right") {
    const swipedCard = data[index];
    // handle accept/reject here
    // if (direction === "right") console.log("ACCEPT", swipedCard);
    // else console.log("REJECT", swipedCard);

    position.setValue({ x: 0, y: 0 });
    setIndex((prev) => prev + 1);
  }

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onPanResponderMove: (_, gesture) => {
        position.setValue({ x: gesture.dx, y: gesture.dy });
      },
      onPanResponderRelease: (_, gesture) => {
        if (gesture.dx > swipeThreshold) forceSwipe("right");
        else if (gesture.dx < -swipeThreshold) forceSwipe("left");
        else resetPosition();
      },
    })
  ).current;

  if (index >= data.length) {
    return (
      <SafeAreaProvider>
        <View style={styles.doneContainer}>
          <Text style={styles.doneTitle}>No more cards</Text>
          <Text style={styles.doneText}>Add more data to keep swiping.</Text>
        </View>
      </SafeAreaProvider>
    );
  }

  const topCard = data[index];
  const nextCard = data[index + 1];

  return (
    <SafeAreaProvider>
      <View style={[styles.container, { height }]}>
        {/* EDGE GLOWS (behind the card) */}
        <Animated.View
          pointerEvents="none"
          style={[
            styles.edgeGlow,
            styles.leftEdge,
            { opacity: leftGlowOpacity },
          ]}
        />
        <Animated.View
          pointerEvents="none"
          style={[
            styles.edgeGlow,
            styles.rightEdge,
            { opacity: rightGlowOpacity },
          ]}
        />

        {/* Next card */}
        {nextCard ? (
          <Animated.View style={[styles.card, { transform: [{ scale: nextScale }] }]}>
            <CardContent card={nextCard} />
          </Animated.View>
        ) : null}

        {/* Top draggable card */}
        <Animated.View
          {...panResponder.panHandlers}
          style={[
            styles.card,
            {
              transform: [
                { translateX: position.x },
                { translateY: position.y },
                { rotate },
              ],
            },
          ]}
        >
          <CardContent card={topCard} />
        </Animated.View>
      </View>
    </SafeAreaProvider>
  );
}

function CardContent({ card }: { card: Card }) {
  return (
    <View style={{ flex: 1 }}>
      <Text style={styles.title}>{card.title}</Text>
      <Text style={styles.description}>{card.description}</Text>

      {/* Image placeholder under description */}
      <View style={styles.imagePlaceholder}>
        <Text style={styles.imagePlaceholderText}>Image Placeholder</Text>
      </View>
    </View>
  );
}

const EDGE_WIDTH = 64;

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: "center", alignItems: "center" },

  // Glowing edge panels
  edgeGlow: {
    position: "absolute",
    top: 0,
    bottom: 0,
    width: EDGE_WIDTH,
    zIndex: 5,
  },
  leftEdge: {
    left: 0,
    backgroundColor: "rgba(255, 0, 0, 0.25)",
    // “glow” feel via a subtle border; simple + works cross-platform
    borderRightWidth: 2,
    borderRightColor: "rgba(255, 0, 0, 0.55)",
  },
  rightEdge: {
    right: 0,
    backgroundColor: "rgba(0, 200, 0, 0.25)",
    borderLeftWidth: 2,
    borderLeftColor: "rgba(0, 200, 0, 0.55)",
  },

  card: {
    position: "absolute",
    width: "90%",
    height: "78%",
    borderRadius: 22,
    padding: 18,
    borderWidth: 1,
    backgroundColor: "white",
    zIndex: 2,
  },

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

  doneContainer: { flex: 1, justifyContent: "center", alignItems: "center", padding: 20 },
  doneTitle: { fontSize: 24, fontWeight: "800", marginBottom: 8 },
  doneText: { fontSize: 16, opacity: 0.8, textAlign: "center" },
});
