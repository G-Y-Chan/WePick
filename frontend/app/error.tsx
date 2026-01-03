import { View, Text, Button, StyleSheet } from 'react-native';
import { router } from 'expo-router';
import { useLocalSearchParams } from 'expo-router';

export default function MyPopupModal() {
    const { errorMessage } = useLocalSearchParams();
    return (
        <View style={styles.container}>
            <Text>Error: {errorMessage}</Text>
            <Button
            title="Close Modal"
            onPress={() => {
                // Dismiss the modal using router.back()
                router.back();
            }}
            />
        </View>
    );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
    backgroundColor: 'white', // Ensure the background is solid
  },
});
