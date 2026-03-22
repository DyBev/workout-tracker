import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import * as exerciseStorage from '../services/exerciseStorage';
import * as exerciseApi from '../services/exerciseApi';
import type {
  SavedExercise,
  SavedExerciseAction,
  SavedExerciseState,
  ExerciseContextValue,
} from '../types/workout';

function generateId(): string {
  return `${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

const initialState: SavedExerciseState = {
  savedExercises: [],
  isLoading: true,
};

function exerciseReducer(
  state: SavedExerciseState,
  action: SavedExerciseAction,
): SavedExerciseState {
  switch (action.type) {
    case 'LOAD_EXERCISES':
      return { ...state, savedExercises: action.exercises, isLoading: false };
    case 'ADD_EXERCISE':
      return {
        ...state,
        savedExercises: [...state.savedExercises, action.exercise],
      };
    case 'UPDATE_EXERCISE':
      return {
        ...state,
        savedExercises: state.savedExercises.map((e) =>
          e.savedExerciseId === action.exercise.savedExerciseId
            ? action.exercise
            : e,
        ),
      };
    case 'ARCHIVE_EXERCISE':
      return {
        ...state,
        savedExercises: state.savedExercises.map((e) =>
          e.savedExerciseId === action.savedExerciseId
            ? { ...e, archivedAt: action.archivedAt, updatedAt: action.archivedAt }
            : e,
        ),
      };
    case 'RESTORE_EXERCISE':
      return {
        ...state,
        savedExercises: state.savedExercises.map((e) =>
          e.savedExerciseId === action.savedExerciseId
            ? { ...e, archivedAt: null, updatedAt: new Date().toISOString() }
            : e,
        ),
      };
    default:
      return state;
  }
}

const ExerciseContext = createContext<ExerciseContextValue | null>(null);

interface ExerciseProviderProps {
  children: React.ReactNode;
}

export function ExerciseProvider({ children }: ExerciseProviderProps) {
  const [state, dispatch] = useReducer(exerciseReducer, initialState);

  useEffect(() => {
    async function restore() {
      try {
        // Try API first, fall back to local storage
        const apiResult = await exerciseApi.readExercises();
        if (apiResult.exercises.length > 0) {
          dispatch({ type: 'LOAD_EXERCISES', exercises: apiResult.exercises });
          // Update local cache
          await exerciseStorage.setSavedExercises(apiResult.exercises);
          return;
        }
      } catch {
        // API unavailable — fall through to local storage
      }

      try {
        const exercises = await exerciseStorage.getSavedExercises();
        dispatch({ type: 'LOAD_EXERCISES', exercises });
      } catch {
        dispatch({ type: 'LOAD_EXERCISES', exercises: [] });
      }
    }
    restore();
  }, []);

  const saveExercise = useCallback(
    (name: string): SavedExercise => {
      const now = new Date().toISOString();
      const exercise: SavedExercise = {
        savedExerciseId: generateId(),
        name,
        note: '',
        tags: [],
        archivedAt: null,
        createdAt: now,
        updatedAt: now,
      };
      dispatch({ type: 'ADD_EXERCISE', exercise });
      exerciseStorage.saveExercise(exercise).catch(() => {});
      exerciseApi.saveExercises(exercise).catch(() => {});
      return exercise;
    },
    [],
  );

  const updateExercise = useCallback(
    async (exercise: SavedExercise): Promise<void> => {
      const updated = { ...exercise, updatedAt: new Date().toISOString() };
      dispatch({ type: 'UPDATE_EXERCISE', exercise: updated });
      await exerciseStorage.updateExercise(updated);
      exerciseApi.saveExercises(updated).catch(() => {});
    },
    [],
  );

  const archiveExercise = useCallback(
    async (savedExerciseId: string): Promise<void> => {
      const now = new Date().toISOString();
      dispatch({ type: 'ARCHIVE_EXERCISE', savedExerciseId, archivedAt: now });
      // Update local storage — find the exercise and set archivedAt
      const exercises = await exerciseStorage.getSavedExercises();
      const exercise = exercises.find((e) => e.savedExerciseId === savedExerciseId);
      if (exercise) {
        await exerciseStorage.updateExercise({ ...exercise, archivedAt: now, updatedAt: now });
      }
      exerciseApi.archiveExercise(savedExerciseId, true).catch(() => {});
    },
    [],
  );

  const restoreExercise = useCallback(
    async (savedExerciseId: string): Promise<void> => {
      const now = new Date().toISOString();
      dispatch({ type: 'RESTORE_EXERCISE', savedExerciseId });
      // Update local storage — find the exercise and clear archivedAt
      const exercises = await exerciseStorage.getSavedExercises();
      const exercise = exercises.find((e) => e.savedExerciseId === savedExerciseId);
      if (exercise) {
        await exerciseStorage.updateExercise({ ...exercise, archivedAt: null, updatedAt: now });
      }
      exerciseApi.archiveExercise(savedExerciseId, false).catch(() => {});
    },
    [],
  );

  const getById = useCallback(
    (savedExerciseId: string | null): SavedExercise | undefined => {
      if (!savedExerciseId) return undefined;
      return state.savedExercises.find(
        (e) => e.savedExerciseId === savedExerciseId,
      );
    },
    [state.savedExercises],
  );

  // Active exercises only (not archived)
  const savedExercises = useMemo(
    () => state.savedExercises.filter((e) => e.archivedAt === null),
    [state.savedExercises],
  );

  const value = useMemo<ExerciseContextValue>(
    () => ({
      savedExercises,
      allSavedExercises: state.savedExercises,
      isLoading: state.isLoading,
      saveExercise,
      updateExercise,
      archiveExercise,
      restoreExercise,
      getById,
    }),
    [
      savedExercises,
      state.savedExercises,
      state.isLoading,
      saveExercise,
      updateExercise,
      archiveExercise,
      restoreExercise,
      getById,
    ],
  );

  return (
    <ExerciseContext.Provider value={value}>
      {children}
    </ExerciseContext.Provider>
  );
}

export function useExercise(): ExerciseContextValue {
  const context = useContext(ExerciseContext);
  if (!context) {
    throw new Error('useExercise must be used within an ExerciseProvider');
  }
  return context;
}
