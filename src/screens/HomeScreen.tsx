import React, { useCallback } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { Button } from '../components';
import { useAuth } from '../contexts/AuthContext';
import { useWorkout } from '../contexts/WorkoutContext';
import { colors } from '../constants/colors';
import type { AppStackParamList } from '../types/workout';

type HomeScreenProps = NativeStackScreenProps<AppStackParamList, 'Home'>;

export function HomeScreen({ navigation }: HomeScreenProps) {
  const { state: authState, signOut } = useAuth();
  const { state: workoutState, startWorkout } = useWorkout();
  const [isSigningOut, setIsSigningOut] = React.useState(false);

  const handleSignOut = useCallback(async () => {
    setIsSigningOut(true);
    try {
      await signOut();
    } catch {
      setIsSigningOut(false);
    }
  }, [signOut]);

  const handleStartWorkout = useCallback(() => {
    startWorkout();
    navigation.navigate('ActiveWorkout');
  }, [startWorkout, navigation]);

  const handleResumeWorkout = useCallback(() => {
    navigation.navigate('ActiveWorkout');
  }, [navigation]);

  const handleViewHistory = useCallback(() => {
    navigation.navigate('WorkoutHistory');
  }, [navigation]);

  const hasActiveWorkout = workoutState.activeWorkout !== null;

  return (
    <View style={styles.container}>
      <Text style={styles.title} accessibilityRole="header">
        Home
      </Text>

      <Text style={styles.welcome}>
        Welcome{authState.user?.email ? `, ${authState.user.email}` : ''}
      </Text>

      <View style={styles.actions}>
        {hasActiveWorkout ? (
          <Button
            title="Resume Workout"
            variant="primary"
            onPress={handleResumeWorkout}
            containerStyle={styles.actionButton}
          />
        ) : (
          <Button
            title="Start Workout"
            variant="primary"
            onPress={handleStartWorkout}
            containerStyle={styles.actionButton}
          />
        )}

        <Button
          title="Workout History"
          variant="secondary"
          onPress={handleViewHistory}
          containerStyle={styles.actionButton}
        />
      </View>

      <Button
        title="Sign Out"
        variant="link"
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
    color: colors.primary.greyDark,
    marginBottom: 32,
  },
  actions: {
    width: '100%',
    maxWidth: 320,
  },
  actionButton: {
    width: '100%',
    marginBottom: 12,
  },
  signOutButton: {
    marginTop: 24,
  },
});
