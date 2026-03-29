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
  failedSyncIds: new Set(),
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
    case 'MARK_SYNCED':
      return {
        ...state,
        savedExercises: state.savedExercises.map((e) =>
          action.savedExerciseIds.includes(e.savedExerciseId)
            ? { ...e, syncStatus: 'synced' as const }
            : e,
        ),
      };
    case 'MARK_SYNC_FAILED': {
      const nextFailed = new Set(state.failedSyncIds);
      for (const id of action.savedExerciseIds) {
        nextFailed.add(id);
      }
      return { ...state, failedSyncIds: nextFailed };
    }
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

  const syncExercises = useCallback(
    async (exercises?: SavedExercise[]) => {
      const toSync =
        exercises ??
        state.savedExercises.filter(
          (e) =>
            e.syncStatus !== 'synced' &&
            !state.failedSyncIds.has(e.savedExerciseId),
        );

      if (toSync.length === 0) return;

      const result = await exerciseApi.saveExercises(toSync);

      const ids = result.results
        .filter((e) => e.status === 'saved')
        .map((e) => e.savedExerciseId);

      if (result.success) {
        dispatch({ type: 'MARK_SYNCED', savedExerciseIds: ids });
        await exerciseStorage.deletePendingExercises(ids);

        // Update local saved cache to reflect synced status
        const cached = await exerciseStorage.getSavedExercises();
        const idSet = new Set(ids);
        const updatedCache = cached.map((e) =>
          idSet.has(e.savedExerciseId) ? { ...e, syncStatus: 'synced' as const } : e,
        );
        await exerciseStorage.setSavedExercises(updatedCache);
      } else {
        dispatch({ type: 'MARK_SYNC_FAILED', savedExerciseIds: ids });
      }
    },
    [state.savedExercises, state.failedSyncIds],
  );

  useEffect(() => {
    async function restore() {
      try {
        const [pending, apiResult] = await Promise.all([
          exerciseStorage.getPendingExercises(),
          exerciseApi.readExercises(),
        ]);

        const apiMap = new Map<string, SavedExercise>(
          apiResult.exercises.map((e) => [e.savedExerciseId, { ...e, syncStatus: 'synced' as const }]),
        );

        for (const e of pending) {
          if (!apiMap.has(e.savedExerciseId)) {
            apiMap.set(e.savedExerciseId, e);
          }
        }

        const merged = Array.from(apiMap.values());
        dispatch({ type: 'LOAD_EXERCISES', exercises: merged });

        // Update local cache with the merged result
        await exerciseStorage.setSavedExercises(merged);

        if (pending.length > 0) {
          syncExercises(pending).catch(() => {});
        }
      } catch {
        // API unavailable — fall back to local storage
        try {
          const exercises = await exerciseStorage.getSavedExercises();
          dispatch({ type: 'LOAD_EXERCISES', exercises });
        } catch {
          dispatch({ type: 'LOAD_EXERCISES', exercises: [] });
        }
      }
    }
    restore();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
        syncStatus: 'pending',
      };
      dispatch({ type: 'ADD_EXERCISE', exercise });
      exerciseStorage.savePendingExercise(exercise).catch(() => {});
      return exercise;
    },
    [],
  );

  const updateExercise = useCallback(
    async (exercise: SavedExercise): Promise<void> => {
      const updated: SavedExercise = {
        ...exercise,
        updatedAt: new Date().toISOString(),
        syncStatus: 'pending',
      };
      dispatch({ type: 'UPDATE_EXERCISE', exercise: updated });
      await exerciseStorage.savePendingExercise(updated);
    },
    [],
  );

  const archiveExercise = useCallback(
    async (savedExerciseId: string): Promise<void> => {
      const now = new Date().toISOString();
      dispatch({ type: 'ARCHIVE_EXERCISE', savedExerciseId, archivedAt: now });

      const exercises = await exerciseStorage.getSavedExercises();
      const exercise = exercises.find((e) => e.savedExerciseId === savedExerciseId);
      if (exercise) {
        const updated: SavedExercise = { ...exercise, archivedAt: now, updatedAt: now, syncStatus: 'synced' };
        await exerciseStorage.updateExercise(updated);
      }
      exerciseApi.archiveExercise(savedExerciseId, true).catch(() => {});
    },
    [],
  );

  const restoreExercise = useCallback(
    async (savedExerciseId: string): Promise<void> => {
      const now = new Date().toISOString();
      dispatch({ type: 'RESTORE_EXERCISE', savedExerciseId });

      const exercises = await exerciseStorage.getSavedExercises();
      const exercise = exercises.find((e) => e.savedExerciseId === savedExerciseId);
      if (exercise) {
        const updated: SavedExercise = { ...exercise, archivedAt: null, updatedAt: now, syncStatus: 'synced' };
        await exerciseStorage.updateExercise(updated);
      }
      exerciseApi.archiveExercise(savedExerciseId, false).catch(() => {});
    },
    [],
  );

  const getById = useCallback(
    (savedExerciseId?: string): SavedExercise | undefined => {
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
      syncExercises,
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
      syncExercises,
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
