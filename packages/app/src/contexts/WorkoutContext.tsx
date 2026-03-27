import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import * as workoutStorage from '../services/workoutStorage';
import * as workoutApi from '../services/workoutApi';
import { useAuth } from './AuthContext';
import { useExercise } from './ExerciseContext';
import type {
  Workout,
  WorkoutAction,
  WorkoutContextValue,
  WorkoutExercise,
  WorkoutSet,
  WorkoutState,
} from '../types/workout';
import { useNavigation } from '@react-navigation/native';

function generateId(): string {
  return `${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

const initialState: WorkoutState = {
  activeWorkout: null,
  history: [],
  isLoading: true,
  failedSyncIds: new Set(),
  nextSk: "",
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
	case 'SET_NEXT_SK': {
		return {
			...state,
			nextSk: action.sk,
		};
	}


    default:
      return state;
  }
}

const WorkoutContext = createContext<WorkoutContextValue | null>(null);

interface WorkoutProviderProps {
  children: React.ReactNode;
}

export function WorkoutProvider({ children }: WorkoutProviderProps) {
  const [state, dispatch] = useReducer(workoutReducer, initialState);
  const { state: authState } = useAuth();
  const { syncExercises } = useExercise();
  const navigation = useNavigation();

  useEffect(() => {
    async function restore() {
      try {
        const [active, pending, workouts] = await Promise.all([
          workoutStorage.getActiveWorkout(),
          workoutStorage.getPendingWorkouts(),
          workoutApi.getWorkouts(""),
        ]);
		if (workouts.nextSk) {
			dispatch({ type: 'SET_NEXT_SK', sk: workouts.nextSk });
		}

        const apiMap = new Map(workouts.workouts.map((w) => [w.workoutId, w]));

		if (active && !apiMap.has(active.workoutId)) {
          dispatch({ type: 'START_WORKOUT', workout: active });
        } else {
			workoutStorage.clearActiveWorkout();
		}

        for (const w of pending) {
          if (!apiMap.has(w.workoutId)) {
            apiMap.set(w.workoutId, w);
          }
        }
        const merged = Array.from(apiMap.values()).sort(
          (a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt),
        );
        dispatch({ type: 'LOAD_HISTORY', workouts: merged });

        if (pending.length > 0) {
          syncWorkouts(pending).catch(() => {});
        }
      } catch {
        dispatch({ type: 'SET_LOADING', isLoading: false });
      }
    }

    restore();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (state.activeWorkout) {
      workoutStorage.saveActiveWorkout(state.activeWorkout);
    }
  }, [state.activeWorkout]);

  const startWorkout = useCallback(() => {
    const now = new Date().toISOString();
    const workoutId = generateId();

    const workout: Workout = {
      workoutId,
      templateId: null,
      startedAt: now,
      completedAt: null,
      notes: '',
      tags: [],
      bodyWeight: 0,
      exercises: [],
      createdAt: now,
      updatedAt: now,
      syncStatus: 'pending',
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
    (name: string, savedExerciseId?: string | null) => {
      updateActiveWorkout((workout) => {
        const exerciseId = generateId();
        const order = workout.exercises.length + 1;
        const newExercise: WorkoutExercise = {
          exerciseId,
          savedExerciseId: savedExerciseId ?? null,
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

  const updateExerciseSavedId = useCallback(
    (
      exerciseId: string,
      savedExerciseId: string,
    ) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises.map((exercise) => {
          console.log(exerciseId, exercise.exerciseId);
          if (exercise.exerciseId !== exerciseId) return exercise;
          return { ...exercise, savedExerciseId };
        });
        console.log(exercises);
        return { ...workout, exercises };
      });
    },
    [updateActiveWorkout],
  );

  const updateExerciseNote = useCallback(
    (exerciseId: string, note: string | null) => {
      updateActiveWorkout((workout) => {
        const exercises = workout.exercises.map((exercise) =>
          exercise.exerciseId === exerciseId ? { ...exercise, note: note ?? undefined } : exercise,
        );
        const updated = { ...workout, exercises };
        // persist immediately so a subsequent save uses the latest value
        workoutStorage.saveActiveWorkout(updated).catch(() => {});
        return updated;
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
    (bodyWeight: number) => {
      updateActiveWorkout((workout) => ({ ...workout, bodyWeight }));
    },
    [updateActiveWorkout],
  );

  const updateExerciseName = useCallback(
    (exerciseId: string, name: string, savedExerciseId: string | null) => {
      updateActiveWorkout((workout) => ({
        ...workout,
        exercises: workout.exercises.map((e) =>
          e.exerciseId === exerciseId ? { ...e, name, savedExerciseId } : e,
        ),
      }));
    },
    [updateActiveWorkout],
  );

  const syncWorkouts = useCallback(
    async (workouts?: Workout[]) => {
      const toSync =
        workouts ??
        state.history.filter(
          (w) =>
            w.syncStatus !== 'synced' &&
            !state.failedSyncIds.has(w.workoutId),
        );

      if (toSync.length === 0) return;

      const payload = toSync.length === 1 ? toSync[0] : toSync;
      const result = await workoutApi.saveWorkouts(payload);

      const ids = result.results.filter((w) => w.status === 'saved').map((w) => w.workoutId);

      if (result.success) {
        dispatch({ type: 'MARK_SYNCED', workoutIds: ids });
        await workoutStorage.deletePendingWorkouts(ids);
      } else {
        dispatch({ type: 'MARK_SYNC_FAILED', workoutIds: ids });
      }
    },
    [state.history, state.failedSyncIds],
  );

  const completeWorkout = useCallback(async () => {
    if (!state.activeWorkout) return;
    // yield to the event loop so any pending state updates (e.g. note edits)
    // have a chance to flush before we snapshot the active workout.
    await Promise.resolve();

    const now = new Date().toISOString();
    const completed: Workout = {
      ...state.activeWorkout,
      completedAt: now,
      updatedAt: now,
      syncStatus: 'pending',
    };

    await workoutStorage.savePendingWorkout(completed);
    await workoutStorage.clearActiveWorkout();
    dispatch({ type: 'COMPLETE_WORKOUT', workout: completed });

    syncWorkouts([completed]).catch(() => {});
    syncExercises().catch(() => {});
  }, [state.activeWorkout, syncWorkouts, syncExercises]);

  const discardWorkout = useCallback(async () => {
    await workoutStorage.clearActiveWorkout();
    dispatch({ type: 'DISCARD_WORKOUT' });
    navigation.navigate('Home' as never);
  }, [navigation]);

  const value = useMemo<WorkoutContextValue>(
    () => ({
      state,
      startWorkout,
      addExercise,
      removeExercise,
      addSet,
      removeSet,
      updateSet,
      updateExerciseName,
      updateExerciseSavedId,
      updateNotes,
      updateTags,
      updateBodyWeight,
      updateExerciseNote,
      completeWorkout,
      discardWorkout,
      syncWorkouts,
    }),
    [
      state,
      startWorkout,
      addExercise,
      removeExercise,
      addSet,
      removeSet,
      updateSet,
      updateExerciseName,
      updateExerciseSavedId,
      updateNotes,
      updateTags,
      updateBodyWeight,
      updateExerciseNote,
      completeWorkout,
      discardWorkout,
      syncWorkouts,
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
