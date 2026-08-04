import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Animated, PanResponder } from "react-native";
import { Card, SwipeDirection } from "./types";
import { SWIPE_OUT_DURATION_MS, SWIPE_THRESHOLD_RATIO } from "./constants";

type Params = {
  data: Card[];
  width: number;
  onSwipe?: (card: Card, direction: SwipeDirection) => void;
};

export function useSwipeDeck({ data, width, onSwipe }: Params) {
  const [index, setIndex] = useState(0);

  const indexRef = useRef(index);
  useEffect(() => {
    indexRef.current = index;
  }, [index]);

  const dataRef = useRef(data);
  useEffect(() => {
    dataRef.current = data;
  }, [data]);

  const position = useRef(new Animated.ValueXY({ x: 0, y: 0 })).current;

  const swipeThreshold = useMemo(
    () => width * SWIPE_THRESHOLD_RATIO,
    [width]
  );

  const rotate = useMemo(
    () =>
      position.x.interpolate({
        inputRange: [-width, 0, width],
        outputRange: ["-12deg", "0deg", "12deg"],
      }),
    [position.x, width]
  );

  const leftGlowOpacity = useMemo(
    () =>
      position.x.interpolate({
        inputRange: [-swipeThreshold * 1.8, -swipeThreshold, 0],
        outputRange: [0.35, 1, 0],
        extrapolate: "clamp",
      }),
    [position.x, swipeThreshold]
  );

  const rightGlowOpacity = useMemo(
    () =>
      position.x.interpolate({
        inputRange: [0, swipeThreshold, swipeThreshold * 1.8],
        outputRange: [0, 0.35, 1],
        extrapolate: "clamp",
      }),
    [position.x, swipeThreshold]
  );

  const nextScale = useMemo(
    () =>
      position.x.interpolate({
        inputRange: [-width, 0, width],
        outputRange: [0.98, 0.95, 0.98],
        extrapolate: "clamp",
      }),
    [position.x, width]
  );

  const resetPosition = useCallback(() => {
    Animated.spring(position, {
      toValue: { x: 0, y: 0 },
      useNativeDriver: false,
      friction: 6,
    }).start();
  }, [position]);

  const onSwipeComplete = useCallback(
    (direction: SwipeDirection) => {
      const i = indexRef.current;
      const swipedCard = dataRef.current[i];

      if (swipedCard && onSwipe) onSwipe(swipedCard, direction);

      position.setValue({ x: 0, y: 0 });
      setIndex((prev) => prev + 1);
    },
    [onSwipe, position]
  );

  const forceSwipe = useCallback(
    (direction: SwipeDirection) => {
      const x = direction === "right" ? width * 1.2 : -width * 1.2;

      Animated.timing(position, {
        toValue: { x, y: 0 },
        duration: SWIPE_OUT_DURATION_MS,
        useNativeDriver: false,
      }).start(() => onSwipeComplete(direction));
    },
    [width, position, onSwipeComplete]
  );

  const panResponder = useMemo(
    () =>
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
      }),
    [position, swipeThreshold, forceSwipe, resetPosition]
  );

  const done = index >= data.length;
  const topCard = data[index];
  const nextCard = data[index + 1];

  // 1. Grab the next 2 cards to silently mount as a background image buffer
  const bufferCards = data.slice(index + 2, index + 4);

  return {
    index,
    done,
    topCard,
    nextCard,
    bufferCards, // 2. Export bufferCards so <SwipeDeck /> can render them hidden
    panHandlers: panResponder.panHandlers,
    position,
    rotate,
    leftGlowOpacity,
    rightGlowOpacity,
    nextScale,
    forceSwipe,
    resetPosition,
  };
}
