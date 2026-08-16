import React from "react";
import { View, Text, StyleSheet, Animated } from "react-native";
import { SwipeCard } from "./SwipeCard";
import { EDGE_WIDTH } from "./constants";
import { Card } from "./types";

type Props = {
  height: number;
  topCard?: Card;
  nextCard?: Card;
  bufferCards?: Card[];
  panHandlers: any;
  position: Animated.ValueXY;
  rotate: any;
  leftGlowOpacity: any;
  rightGlowOpacity: any;
  nextScale: any;
};

export function SwipeDeck(props: Props) {
  const {
    height,
    topCard,
    nextCard,
    bufferCards = [],
    panHandlers,
    position,
    rotate,
    leftGlowOpacity,
    rightGlowOpacity,
    nextScale,
  } = props;

  return (
    <View style={[styles.container, { height }]}>
      <Animated.View
        pointerEvents="none"
        style={[styles.edgeGlow, styles.leftEdge, { opacity: leftGlowOpacity }]}
      />
      <Animated.View
        pointerEvents="none"
        style={[styles.edgeGlow, styles.rightEdge, { opacity: rightGlowOpacity }]}
      />

      {bufferCards.map((card) => (
        <View
          key={`buffer-${card.id}`}
          pointerEvents="none"
          style={[styles.card, styles.hiddenBufferCard]}
        >
          <SwipeCard card={card} />
        </View>
      ))}

      {nextCard ? (
        <Animated.View
          key={`next-${nextCard.id}`}
          style={[styles.card, { transform: [{ scale: nextScale }] }]}
        >
          <SwipeCard card={nextCard} />
        </Animated.View>
      ) : null}

      {topCard ? (
        <Animated.View
          key={`top-${topCard.id}`}
          {...panHandlers}
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
          <SwipeCard card={topCard} />
        </Animated.View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: "center", alignItems: "center" },
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
  hiddenBufferCard: {
    opacity: 0.01,
    zIndex: -1,
  },
});
