import React, { useEffect, useState, useCallback } from "react";
import { View, Text, StyleSheet, useWindowDimensions } from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";

import { useLocalSearchParams, useRouter } from "expo-router";

import { Card, SwipeDirection } from "../../src/swipe/types";
import { useSwipeDeck } from "../../src/swipe/useSwipeDeck";
import { SwipeDeck } from "../../src/swipe/SwipeDeck";

import { Message } from "@/services/types";
import { getCardData } from "@/services/api/http";

import { usePlaces } from "@/src/context/places-context";
import { useRoom } from "@/src/context/room-context";

export default function SwipeScreen() {
  const [data, setData] = useState<Card[]>([]);

  const { width, height } = useWindowDimensions();

  const { roomCode } = useLocalSearchParams<{ roomCode: string }>();

  const router = useRouter();

  const { setPlaces } = usePlaces();

  const { send, lastMessage } = useRoom();

  // React to server events
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.Header !== "MAJORITY_FOUND") {
      return;
    }

    const voteID = lastMessage.VoteObj?.Id;

    if (!voteID) return;

    router.push({
      pathname: "/results",
      params: { id: voteID },
    });
  }, [lastMessage, router]);

  // Load cards
  useEffect(() => {
    let cancelled = false;

    async function load() {
      const location = "PLACEHOLDER_LOCATION";

      const res = await getCardData(location);

      if (!cancelled) {
        const cards = res || [];

        setData(cards);

        setPlaces(cards);
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [setPlaces]);

  // Send vote event
  const handleSwipe = useCallback(
    (card: Card, direction: SwipeDirection) => {
      const result =
        direction === "right"
          ? "ACCEPT"
          : "REJECT";

      const voteEventMessage: Message = {
        Header: "VOTE_EVENT",

        VoteObj: {
          Id: card.id,
          Result: result,
          Room: roomCode,
        },
      };

      send(voteEventMessage);
    },
    [send, roomCode]
  );

  const deck = useSwipeDeck({
    data,
    width,
    onSwipe: handleSwipe,
  });

  if (deck.done) {
    return (
      <SafeAreaProvider>
        <View style={styles.doneContainer}>
          <Text style={styles.doneTitle}>
            No more cards
          </Text>

          <Text style={styles.doneText}>
            Add more data to keep swiping.
          </Text>
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
  doneContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    padding: 20,
  },

  doneTitle: {
    fontSize: 24,
    fontWeight: "800",
    marginBottom: 8,
  },

  doneText: {
    fontSize: 16,
    opacity: 0.8,
    textAlign: "center",
  },
});
