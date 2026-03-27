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
  savedExerciseId: string | null;
  name: string;
  order: number;
  sets: WorkoutSet[];
  note?: string;
}

export type SyncStatus = 'pending' | 'synced';

export interface Workout {
  workoutId: string;
  templateId: string | null;
  startedAt: string;
  completedAt: string | null;
  notes: string;
  tags: string[];
  bodyWeight: number;
  exercises: WorkoutExercise[];
  createdAt: string;
  updatedAt: string;
  syncStatus: SyncStatus;
}

export interface SavedExercise {
  savedExerciseId: string;
  name: string;
  note: string;
  tags: string[];
  archivedAt: string | null;
  createdAt: string;
  updatedAt: string;
  syncStatus: SyncStatus;
}

export type AppStackParamList = {
  Home: undefined;
  ActiveWorkout: undefined;
  WorkoutHistory: undefined;
  WorkoutSummary: { workoutId: string };
  SavedExercises: undefined;
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

export type SavedExercisesScreenProps = NativeStackScreenProps<
  AppStackParamList,
  'SavedExercises'
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
  | { type: 'SET_NEXT_SK'; sk: string }
  | { type: 'MARK_SYNC_FAILED'; workoutIds: string[] };


export interface WorkoutContextValue {
  state: WorkoutState;
  startWorkout: () => void;
  addExercise: (name: string, savedExerciseId?: string | null) => void;
  removeExercise: (exerciseId: string) => void;
  addSet: (exerciseId: string) => void;
  removeSet: (exerciseId: string, setId: string) => void;
  updateSet: (
    exerciseId: string,
    setId: string,
    field: 'reps' | 'weight',
    value: number | null,
  ) => void;
  updateExerciseName: (
    exerciseId: string,
    name: string,
    savedExerciseId: string | null,
  ) => void;
  updateExerciseSavedId: (
    exerciseId: string,
    savedExerciseId: string,
  ) => void;
  updateExerciseNote: (exerciseId: string, note: string | null) => void;
  updateNotes: (notes: string) => void;
  updateTags: (tags: string[]) => void;
  updateBodyWeight: (weight: number) => void;
  completeWorkout: () => Promise<void>;
  discardWorkout: () => Promise<void>;
  syncWorkouts: (workouts?: Workout[]) => Promise<void>;
}

export interface SavedExerciseState {
  savedExercises: SavedExercise[];
  isLoading: boolean;
  failedSyncIds: Set<string>;
}

export type SavedExerciseAction =
  | { type: 'LOAD_EXERCISES'; exercises: SavedExercise[] }
  | { type: 'ADD_EXERCISE'; exercise: SavedExercise }
  | { type: 'UPDATE_EXERCISE'; exercise: SavedExercise }
  | { type: 'ARCHIVE_EXERCISE'; savedExerciseId: string; archivedAt: string }
  | { type: 'RESTORE_EXERCISE'; savedExerciseId: string }
  | { type: 'MARK_SYNCED'; savedExerciseIds: string[] }
  | { type: 'MARK_SYNC_FAILED'; savedExerciseIds: string[] };

export interface ExerciseContextValue {
  savedExercises: SavedExercise[];
  allSavedExercises: SavedExercise[];
  isLoading: boolean;
  saveExercise: (name: string) => SavedExercise;
  updateExercise: (exercise: SavedExercise) => Promise<void>;
  archiveExercise: (savedExerciseId: string) => Promise<void>;
  restoreExercise: (savedExerciseId: string) => Promise<void>;
  getById: (savedExerciseId: string | null) => SavedExercise | undefined;
  syncExercises: (exercises?: SavedExercise[]) => Promise<void>;
}
