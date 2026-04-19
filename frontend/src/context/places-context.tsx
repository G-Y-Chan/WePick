import React, {
  createContext,
  useContext,
  useMemo,
  useState,
  ReactNode,
  useCallback,
} from "react";
import { Card } from "@/src/swipe/types";

type PlacesContextValue = {
  placesById: Record<string, Card>;
  setPlaces: (places: Card[]) => void;
  getPlaceById: (id: string) => Card | undefined;
  clearPlaces: () => void;
};

const PlacesContext = createContext<PlacesContextValue | undefined>(undefined);

export function PlacesProvider({ children }: { children: ReactNode }) {
  const [placesById, setPlacesById] = useState<Record<string, Card>>({});

  const setPlaces = useCallback((places: Card[]) => {
    const next: Record<string, Card> = {};

    for (const place of places) {
      next[place.id] = place;
    }

    setPlacesById((prev) => ({
      ...prev,
      ...next,
    }));
  }, []);

  const getPlaceById = useCallback(
    (id: string) => {
      return placesById[id];
    },
    [placesById]
  );

  const clearPlaces = useCallback(() => {
    setPlacesById({});
  }, []);

  const value = useMemo(
    () => ({
      placesById,
      setPlaces,
      getPlaceById,
      clearPlaces,
    }),
    [placesById, setPlaces, getPlaceById, clearPlaces]
  );

  return (
    <PlacesContext.Provider value={value}>
      {children}
    </PlacesContext.Provider>
  );
}

export function usePlaces() {
  const ctx = useContext(PlacesContext);

  if (!ctx) {
    throw new Error("usePlaces must be used within a PlacesProvider");
  }

  return ctx;
}