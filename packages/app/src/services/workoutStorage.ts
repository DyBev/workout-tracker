import AsyncStorage from '@react-native-async-storage/async-storage';
import type { Workout } from '../types/workout';

const ACTIVE_WORKOUT_KEY = 'workout:active';
const HISTORY_KEY = 'workout:history';

// ── Active workout ───────────────────────────────────────────────────────────

export async function saveActiveWorkout(workout: Workout): Promise<void> {
  await AsyncStorage.setItem(ACTIVE_WORKOUT_KEY, JSON.stringify(workout));
}

export async function getActiveWorkout(): Promise<Workout | null> {
  const raw = await AsyncStorage.getItem(ACTIVE_WORKOUT_KEY);
  if (!raw) return null;
  return JSON.parse(raw) as Workout;
}

export async function clearActiveWorkout(): Promise<void> {
  await AsyncStorage.removeItem(ACTIVE_WORKOUT_KEY);
}

// ── Workout history ──────────────────────────────────────────────────────────

export async function getWorkoutHistory(): Promise<Workout[]> {
  const raw = await AsyncStorage.getItem(HISTORY_KEY);
  if (!raw) return [];
  return JSON.parse(raw) as Workout[];
}

export async function saveWorkoutToHistory(workout: Workout): Promise<void> {
  const history = await getWorkoutHistory();
  // Prepend newest first
  history.unshift(workout);
  await AsyncStorage.setItem(HISTORY_KEY, JSON.stringify(history));
}

export async function deleteWorkoutFromHistory(
  workoutId: string,
): Promise<void> {
  const history = await getWorkoutHistory();
  const updated = history.filter((w) => w.workoutId !== workoutId);
  await AsyncStorage.setItem(HISTORY_KEY, JSON.stringify(updated));
}
