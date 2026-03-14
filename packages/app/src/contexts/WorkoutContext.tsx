import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import * as workoutStorage from '../services/workoutStorage';
import { useAuth } from './AuthContext';
import type {
  BodyWeight,
  Workout,
  WorkoutAction,
  WorkoutContextValue,
  WorkoutExercise,
  WorkoutSet,
  WorkoutState,
} from '../types/workout';
import { useNavigation } from '@react-navigation/native';

// ── Helpers ──────────────────────────────────────────────────────────────────

function generateId(): string {
  return `${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

function buildSortKey(timestamp: string, workoutId: string): string {
  return `WORKOUT#${timestamp}#${workoutId}`;
}

// ── Reducer ──────────────────────────────────────────────────────────────────

const initialState: WorkoutState = {
  activeWorkout: null,
  history: [],
  isLoading: true,
};

function workoutReducer(
  state: WorkoutState,
  action: WorkoutAction,
): WorkoutState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.isLoading };
    case 'START_WORKOUT':
      return { ...state, activeWorkout: action.workout };
    case 'UPDATE_WORKOUT':
      return { ...state, activeWorkout: action.workout };
    case 'COMPLETE_WORKOUT':
      return {
        ...state,
        activeWorkout: null,
        history: [action.workout, ...state.history],
      };
    case 'DISCARD_WORKOUT':
      return { ...state, activeWorkout: null };
    case 'LOAD_HISTORY':
      return { ...state, history: action.workouts, isLoading: false };
    case 'RESTORE_ACTIVE_WORKOUT':
      return { ...state, activeWorkout: action.workout };
    default:
      return state;
  }
}

// ── Context ──────────────────────────────────────────────────────────────────

const WorkoutContext = createContext<WorkoutContextValue | null>(null);

interface WorkoutProviderProps {
  children: React.ReactNode;
}

export function WorkoutProvider({ children }: WorkoutProviderProps) {
  const [state, dispatch] = useReducer(workoutReducer, initialState);
  const { state: authState } = useAuth();
  const navigation = useNavigation();

  useEffect(() => {
    async function restore() {
      try {
        const [active, history] = await Promise.all([
          workoutStorage.getActiveWorkout(),
          workoutStorage.getWorkoutHistory(),
        ]);
        if (active) {
          dispatch({ type: 'RESTORE_ACTIVE_WORKOUT', workout: active });
        }
        dispatch({ type: 'LOAD_HISTORY', workouts: history });
      } catch {
        dispatch({ type: 'SET_LOADING', isLoading: false });
      }
    }

    restore();
  }, []);

  // Persist active workout on every change
  useEffect(() => {
    if (state.activeWorkout) {
      workoutStorage.saveActiveWorkout(state.activeWorkout);
    }
  }, [state.activeWorkout]);

  const startWorkout = useCallback(() => {
    const now = new Date().toISOString();
    const workoutId = generateId();
    const userId = authState.user?.sub ?? 'local';

    const workout: Workout = {
      userId,
      sk: buildSortKey(now, workoutId),
      workoutId,
      templateId: null,
      startedAt: now,
      completedAt: null,
      notes: '',
      tags: [],
      bodyWeight: null,
      exercises: [],
      createdAt: now,
      updatedAt: now,
    };

    dispatch({ type: 'START_WORKOUT', workout });
  }, [authState.user?.sub]);

  const updateActiveWorkout = useCallback(
    (updater: (workout: Workout) => Workout) => {
      if (!state.activeWorkout) return;
      const updated = updater({
        ...state.activeWorkout,
        updatedAt: new Date().toISOString(),
      });
      dispatch({ type: 'UPDATE_WORKOUT', workout: updated });
    },
    [state.activeWorkout],
  );

  const addExercise = useCallback(
    (name: string) => {
      updateActiveWorkout((workout) => {
        const exerciseId = generateId();
        const order = workout.exercises.length + 1;
        const newExercise: WorkoutExercise = {
          exerciseId,
          name,
          order,
          sets: [],
        };
        return { ...workout, exercises: [...workout.exercises, newExercise] };
      });
    },
    [updateActiveWorkout],
  );

  const removeExercise = useCallback(
    (exerciseId: string) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises
          .filter((e) => e.exerciseId !== exerciseId)
          .map((e, i) => ({ ...e, order: i + 1 }));
        return { ...workout, exercises };
      });
    },
    [updateActiveWorkout],
  );

  const addSet = useCallback(
    (exerciseId: string) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises.map((exercise) => {
          if (exercise.exerciseId !== exerciseId) return exercise;
          const setId = generateId();
          const order = exercise.sets.length + 1;
          const newSet: WorkoutSet = {
            setId,
            order,
            reps: null,
            weight: null,
            weightUnit: 'kg',
          };
          return { ...exercise, sets: [...exercise.sets, newSet] };
        });
        return { ...workout, exercises };
      });
    },
    [updateActiveWorkout],
  );

  const removeSet = useCallback(
    (exerciseId: string, setId: string) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises.map((exercise) => {
          if (exercise.exerciseId !== exerciseId) return exercise;
          const sets = exercise.sets
            .filter((s) => s.setId !== setId)
            .map((s, i) => ({ ...s, order: i + 1 }));
          return { ...exercise, sets };
        });
        return { ...workout, exercises };
      });
    },
    [updateActiveWorkout],
  );

  const updateSet = useCallback(
    (
      exerciseId: string,
      setId: string,
      field: 'reps' | 'weight',
      value: number | null,
    ) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises.map((exercise) => {
          if (exercise.exerciseId !== exerciseId) return exercise;
          const sets = exercise.sets.map((s) => {
            if (s.setId !== setId) return s;
            return { ...s, [field]: value };
          });
          return { ...exercise, sets };
        });
        return { ...workout, exercises };
      });
    },
    [updateActiveWorkout],
  );

  const updateNotes = useCallback(
    (notes: string) => {
      updateActiveWorkout((workout) => ({ ...workout, notes }));
    },
    [updateActiveWorkout],
  );

  const updateTags = useCallback(
    (tags: string[]) => {
      updateActiveWorkout((workout) => ({ ...workout, tags }));
    },
    [updateActiveWorkout],
  );

  const updateBodyWeight = useCallback(
    (bodyWeight: BodyWeight | null) => {
      updateActiveWorkout((workout) => ({ ...workout, bodyWeight }));
    },
    [updateActiveWorkout],
  );

  const completeWorkout = useCallback(async () => {
    if (!state.activeWorkout) return;

    const now = new Date().toISOString();
    const completed: Workout = {
      ...state.activeWorkout,
      completedAt: now,
      updatedAt: now,
    };

    await workoutStorage.saveWorkoutToHistory(completed);
    await workoutStorage.clearActiveWorkout();
    dispatch({ type: 'COMPLETE_WORKOUT', workout: completed });
  }, [state.activeWorkout]);

  const discardWorkout = useCallback(async () => {
    await workoutStorage.clearActiveWorkout();
    dispatch({ type: 'DISCARD_WORKOUT' });
    navigation.navigate('Home');
  }, [navigation]);

  const loadHistory = useCallback(async () => {
    const history = await workoutStorage.getWorkoutHistory();
    dispatch({ type: 'LOAD_HISTORY', workouts: history });
  }, []);

  const value = useMemo<WorkoutContextValue>(
    () => ({
      state,
      startWorkout,
      addExercise,
      removeExercise,
      addSet,
      removeSet,
      updateSet,
      updateNotes,
      updateTags,
      updateBodyWeight,
      completeWorkout,
      discardWorkout,
      loadHistory,
    }),
    [
      state,
      startWorkout,
      addExercise,
      removeExercise,
      addSet,
      removeSet,
      updateSet,
      updateNotes,
      updateTags,
      updateBodyWeight,
      completeWorkout,
      discardWorkout,
      loadHistory,
    ],
  );

  return (
    <WorkoutContext.Provider value={value}>{children}</WorkoutContext.Provider>
  );
}

export function useWorkout(): WorkoutContextValue {
  const context = useContext(WorkoutContext);

  if (!context) {
    throw new Error('useWorkout must be used within a WorkoutProvider');
  }

  return context;
}
