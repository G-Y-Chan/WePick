import React, { useEffect, useState } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";
import {
  ActivityIndicator,
  Button,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { router } from "expo-router";
import { useLocalSearchParams } from 'expo-router';

export default function Room() {
  const handleStartRoom = () => {
      try {
        const response = 'true';
        console.log("room code:", roomCode);
        const status: boolean = response.toLowerCase() === 'true';
        if (status) {
          router.push({pathname: "/swipe"});
        } else {
          const message = "Unable to Start"
          router.push({
            pathname: "/error",
            params: {errorMessage: message },
          })
        }
      } catch (e: unknown) {
        console.error("Error in Starting Room:", e);
        const message = "Internal Server Error"
          router.push({
            pathname: "/error",
            params: {errorMessage: message },
          })
      }
  }

  const { roomCode, host } = useLocalSearchParams();
  const isHost = host === "true";
  return (
    <SafeAreaProvider>
      { isHost ? (
        <>
          <View style={styles.container}>
            <Button title="Start" onPress={handleStartRoom} />
            <Text style={styles.text}>Code to join: {roomCode}</Text>
          </View>
        </>
      ) : (
        <>
          <View style={styles.container}>
            <Text>Waiting for host...</Text>
            <Text style={styles.text}>Code to join: {roomCode}</Text>
          </View>
        </>
      )}
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,                 // take full screen
    justifyContent: "center", // center vertically
    alignItems: "center",     // center horizontally
    paddingHorizontal: 16,
  },
  title: {
    fontSize: 18,
    marginBottom: 8,
    textAlign: "center",
    fontWeight: "600",
  },
  text: {
    fontSize: 16,
    textAlign: "center",
  },
  input: {
    borderWidth: 1,
    borderColor: "#ccc",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
  },
});