import type { NativeStackScreenProps } from '@react-navigation/native-stack';

export interface WorkoutSet {
  setId: string;
  order: number;
  reps: number | null;
  weight: number | null;
  weightUnit: 'kg' | 'lbs';
}

export interface WorkoutExercise {
  exerciseId: string;
  name: string;
  order: number;
  sets: WorkoutSet[];
}

export interface BodyWeight {
  value: number;
  unit: 'kg' | 'lbs';
}

export type SyncStatus = 'pending' | 'synced';

export interface Workout {
  workoutId: string;
  templateId: string | null;
  startedAt: string;
  completedAt: string | null;
  notes: string;
  tags: string[];
  bodyWeight: BodyWeight | null;
  exercises: WorkoutExercise[];
  createdAt: string;
  updatedAt: string;
  syncStatus: SyncStatus;
}

export type AppStackParamList = {
  Home: undefined;
  ActiveWorkout: undefined;
  WorkoutHistory: undefined;
  WorkoutSummary: { workoutId: string };
};

export type ActiveWorkoutScreenProps = NativeStackScreenProps<
  AppStackParamList,
  'ActiveWorkout'
>;

export type WorkoutHistoryScreenProps = NativeStackScreenProps<
  AppStackParamList,
  'WorkoutHistory'
>;

export type WorkoutSummaryScreenProps = NativeStackScreenProps<
  AppStackParamList,
  'WorkoutSummary'
>;

export interface WorkoutState {
  activeWorkout: Workout | null;
  history: Workout[];
  isLoading: boolean;
  failedSyncIds: Set<string>;
  nextSk: string;
}

export type WorkoutAction =
  | { type: 'SET_LOADING'; isLoading: boolean }
  | { type: 'START_WORKOUT'; workout: Workout }
  | { type: 'UPDATE_WORKOUT'; workout: Workout }
  | { type: 'COMPLETE_WORKOUT'; workout: Workout }
  | { type: 'DISCARD_WORKOUT' }
  | { type: 'LOAD_HISTORY'; workouts: Workout[] }
  | { type: 'RESTORE_ACTIVE_WORKOUT'; workout: Workout }
  | { type: 'MARK_SYNCED'; workoutIds: string[] }
  | { type: 'MARK_SYNC_FAILED'; workoutIds: string[] }
  | { type: 'ADD_WORKOUTS'; workouts: Workout[] }
  | { type: 'SET_NEXT_SK'; sk: string };


export interface WorkoutContextValue {
  state: WorkoutState;
  startWorkout: () => void;
  addExercise: (name: string) => void;
  removeExercise: (exerciseId: string) => void;
  addSet: (exerciseId: string) => void;
  removeSet: (exerciseId: string, setId: string) => void;
  updateSet: (
    exerciseId: string,
    setId: string,
    field: 'reps' | 'weight',
    value: number | null,
  ) => void;
  updateNotes: (notes: string) => void;
  updateTags: (tags: string[]) => void;
  updateBodyWeight: (bodyWeight: BodyWeight | null) => void;
  completeWorkout: () => Promise<void>;
  discardWorkout: () => Promise<void>;
  loadHistory: () => Promise<void>;
  syncWorkouts: (workouts?: Workout[]) => Promise<void>;
}
