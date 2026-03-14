import type { NativeStackScreenProps } from '@react-navigation/native-stack';

// ── Set ──────────────────────────────────────────────────────────────────────

export interface WorkoutSet {
  setId: string;
  order: number;
  reps: number | null;
  weight: number | null;
  weightUnit: 'kg' | 'lbs';
}

// ── Exercise ─────────────────────────────────────────────────────────────────

export interface WorkoutExercise {
  exerciseId: string;
  name: string;
  order: number;
  sets: WorkoutSet[];
}

// ── Body-weight map ──────────────────────────────────────────────────────────

export interface BodyWeight {
  value: number;
  unit: 'kg' | 'lbs';
}

// ── Workout (matches DynamoDB single-table item) ─────────────────────────────

export interface Workout {
  /** Partition key — user identifier */
  userId: string;
  /** Sort key — WORKOUT#<timestamp>#<workoutId> */
  sk: string;
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
}

// ── Navigation ───────────────────────────────────────────────────────────────

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

// ── Context state ────────────────────────────────────────────────────────────

export interface WorkoutState {
  activeWorkout: Workout | null;
  history: Workout[];
  isLoading: boolean;
}

export type WorkoutAction =
  | { type: 'SET_LOADING'; isLoading: boolean }
  | { type: 'START_WORKOUT'; workout: Workout }
  | { type: 'UPDATE_WORKOUT'; workout: Workout }
  | { type: 'COMPLETE_WORKOUT'; workout: Workout }
  | { type: 'DISCARD_WORKOUT' }
  | { type: 'LOAD_HISTORY'; workouts: Workout[] }
  | { type: 'RESTORE_ACTIVE_WORKOUT'; workout: Workout };

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
}
