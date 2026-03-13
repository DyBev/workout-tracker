import React, { useCallback } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Button } from '../components';
import { useAuth } from '../contexts/AuthContext';

export function HomeScreen() {
  const { state, signOut } = useAuth();
  const [isSigningOut, setIsSigningOut] = React.useState(false);

  const handleSignOut = useCallback(async () => {
    setIsSigningOut(true);
    try {
      await signOut();
    } catch {
      setIsSigningOut(false);
    }
  }, [signOut]);

  return (
    <View style={styles.container}>
      <Text
        style={styles.title}
        accessibilityRole="header"
      >
        Home
      </Text>

      <Text style={styles.welcome}>
        Welcome{state.user?.email ? `, ${state.user.email}` : ''}
      </Text>

      <Button
        title="Sign Out"
        variant="secondary"
        onPress={handleSignOut}
        loading={isSigningOut}
        containerStyle={styles.signOutButton}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    marginBottom: 8,
  },
  welcome: {
    fontSize: 16,
    color: '#6b7280',
    marginBottom: 32,
  },
  signOutButton: {
    minWidth: 160,
  },
});
