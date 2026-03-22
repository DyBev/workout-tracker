import AsyncStorage from '@react-native-async-storage/async-storage';
import type { SavedExercise } from '../types/workout';

const SAVED_EXERCISES_KEY = 'exercises:saved';

export async function getSavedExercises(): Promise<SavedExercise[]> {
  const raw = await AsyncStorage.getItem(SAVED_EXERCISES_KEY);
  if (!raw) return [];
  return JSON.parse(raw) as SavedExercise[];
}

export async function saveExercise(exercise: SavedExercise): Promise<void> {
  const exercises = await getSavedExercises();
  exercises.push(exercise);
  await AsyncStorage.setItem(SAVED_EXERCISES_KEY, JSON.stringify(exercises));
}

export async function updateExercise(exercise: SavedExercise): Promise<void> {
  const exercises = await getSavedExercises();
  const index = exercises.findIndex(
    (e) => e.savedExerciseId === exercise.savedExerciseId,
  );
  if (index === -1) return;
  exercises[index] = exercise;
  await AsyncStorage.setItem(SAVED_EXERCISES_KEY, JSON.stringify(exercises));
}

export async function setSavedExercises(exercises: SavedExercise[]): Promise<void> {
  await AsyncStorage.setItem(SAVED_EXERCISES_KEY, JSON.stringify(exercises));
}
