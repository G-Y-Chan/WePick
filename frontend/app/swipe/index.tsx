import React, { useMemo, useCallback } from "react";
import { View, Text, StyleSheet, useWindowDimensions } from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { Card, SwipeDirection } from "../../src/swipe/types";
import { useSwipeDeck } from "../../src/swipe/useSwipeDeck";
import { SwipeDeck } from "../../src/swipe/SwipeDeck";

export default function SwipeScreen() {
  const { width, height } = useWindowDimensions();

  const data = useMemo<Card[]>(
    () =>
      Array.from({ length: 10 }, (_, i) => ({
        id: String(i + 1),
        title: `Place ${i + 1}`,
        description: "Placeholder description. Swipe right to accept, left to reject.",
      })),
    []
  );

  const handleSwipe = useCallback((card: Card, direction: SwipeDirection) => {
    // TODAY: console log
    console.log(direction === "right" ? "ACCEPT" : "REJECT", card);

    // LATER: send vote event (websocket)
    // voteClient.sendVote({ roomId, cardId: card.id, direction, ts: Date.now() })
  }, []);

  const deck = useSwipeDeck({ data, width, onSwipe: handleSwipe });

  if (deck.done) {
    return (
      <SafeAreaProvider>
        <View style={styles.doneContainer}>
          <Text style={styles.doneTitle}>No more cards</Text>
          <Text style={styles.doneText}>Add more data to keep swiping.</Text>
        </View>
      </SafeAreaProvider>
    );
  }

  return (
    <SafeAreaProvider>
      <SwipeDeck
        height={height}
        topCard={deck.topCard}
        nextCard={deck.nextCard}
        panHandlers={deck.panHandlers}
        position={deck.position}
        rotate={deck.rotate}
        leftGlowOpacity={deck.leftGlowOpacity}
        rightGlowOpacity={deck.rightGlowOpacity}
        nextScale={deck.nextScale}
      />
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  doneContainer: { flex: 1, justifyContent: "center", alignItems: "center", padding: 20 },
  doneTitle: { fontSize: 24, fontWeight: "800", marginBottom: 8 },
  doneText: { fontSize: 16, opacity: 0.8, textAlign: "center" },
});
