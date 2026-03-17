import AsyncStorage from '@react-native-async-storage/async-storage';
import type { Workout } from '../types/workout';

const ACTIVE_WORKOUT_KEY = 'workout:active';
const PENDING_WORKOUTS_KEY = 'workout:pending';

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

export async function getPendingWorkouts(): Promise<Workout[]> {
  const raw = await AsyncStorage.getItem(PENDING_WORKOUTS_KEY);
  if (!raw) return [];
  return JSON.parse(raw) as Workout[];
}

export async function savePendingWorkout(workout: Workout): Promise<void> {
  const pending = await getPendingWorkouts();
  if (!pending.some((w) => w.workoutId === workout.workoutId)) {
    pending.unshift(workout);
    await AsyncStorage.setItem(PENDING_WORKOUTS_KEY, JSON.stringify(pending));
  }
}

export async function deletePendingWorkouts(workoutIds: string[]): Promise<void> {
  const pending = await getPendingWorkouts();
  const remaining = pending.filter((w) => !workoutIds.includes(w.workoutId));
  await AsyncStorage.setItem(PENDING_WORKOUTS_KEY, JSON.stringify(remaining));
}
