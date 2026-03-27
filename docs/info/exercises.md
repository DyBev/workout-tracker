# Exercises

This document describes the exercise data structures, local storage format, API contracts, and the client-side lifecycle used by the app.

Files used to produce this reference: `packages/app/src/types/workout.ts`, `packages/app/src/services/exerciseStorage.ts`, `packages/app/src/services/exerciseApi.ts`, `packages/app/src/contexts/ExerciseContext.tsx`.

## Data shapes

- **SavedExercise** (persistent exercise template)

```ts
interface SavedExercise {
  savedExerciseId: string; // locally-generated id (string)
  name: string;            // exercise name
  note: string;            // optional note
  tags: string[];          // simple string tags
  archivedAt: string | null; // ISO timestamp when archived, null if active
  createdAt: string;       // ISO timestamp
  updatedAt: string;       // ISO timestamp
  syncStatus: 'pending' | 'synced';
}
```

- **WorkoutExercise** (exercise instance inside a workout)

```ts
interface WorkoutExercise {
  exerciseId: string;        // local id for exercise instance in workout
  savedExerciseId: string | null; // reference to a SavedExercise or null
  name: string;              // displayed name (can come from SavedExercise)
  order: number;             // ordering within workout
  sets: WorkoutSet[];        // array of sets
}

interface WorkoutSet {
  setId: string;
  order: number;
  reps: number | null;
  weight: number | null;
  weightUnit: 'kg' | 'lbs';
}
```

## Local storage

The app uses AsyncStorage to cache both saved exercises and pending changes. Keys used in `packages/app/src/services/exerciseStorage.ts`:

- `exercises:saved` — canonical local cache of saved exercises (array of `SavedExercise`).
- `exercises:pending` — pending changes that need syncing to the API (array of `SavedExercise`).

Important storage helpers in `exerciseStorage.ts`:

- `getSavedExercises(): Promise<SavedExercise[]>` — read saved cache (returns [] if empty).
- `setSavedExercises(exercises: SavedExercise[]): Promise<void>` — overwrite saved cache.
- `saveExercise(exercise: SavedExercise): Promise<void>` — append to saved cache.
- `updateExercise(exercise: SavedExercise): Promise<void>` — replace an existing saved entry by `savedExerciseId`.
- `getPendingExercises(): Promise<SavedExercise[]>` — read pending list.
- `savePendingExercise(exercise: SavedExercise): Promise<void>` — upsert into pending list (new items unshifted to front).
- `deletePendingExercises(savedExerciseIds: string[]): Promise<void>` — remove pending entries by id.

Notes:
- Pending items are stored separately so the UI can show newly-created or updated exercises immediately while the network sync is attempted in background.

## API contract

See `packages/app/src/services/exerciseApi.ts`.

- Read exercises: `GET ${process.env.EXPO_PUBLIC_API_DOMAIN}/exercises/read`
  - Returns a JSON object with `exercises: SavedExercise[]`.

- Save exercises: `POST ${process.env.EXPO_PUBLIC_API_DOMAIN}/exercises/save`
  - Accepts an array of `SavedExercise` JSON objects.
  - Returns status and a result array:

```ts
interface SaveExercisesResult {
  success: boolean;
  status: number;
  results: { savedExerciseId: string; status: 'saved' | 'error' }[];
}
```

- Archive/restore: `PATCH ${process.env.EXPO_PUBLIC_API_DOMAIN}/exercises/archive`
  - Body: `{ savedExerciseId: string, archive: boolean }`
  - Response: `{ success: boolean, status: number }`

All API calls attach Authorization using the current session id token (see `getCurrentSession()` in `packages/app/src/services/auth.ts`). The client falls back to local storage if APIs are unavailable.

## Client lifecycle & sync behavior

Implementation in `packages/app/src/contexts/ExerciseContext.tsx`:

- When the provider mounts it concurrently loads `exercises:pending` and calls the API `readExercises()`.
- Results from the API are merged with pending local items (pending wins for ids that aren't present in the API result). The merged list becomes the canonical in-memory list and is written to `exercises:saved`.
- If pending items exist, the provider attempts to `syncExercises(pending)` in background.

Sync behavior (`syncExercises`):

- Collects exercises to sync (either passed in or all non-synced, non-failed exercises from state).
- Calls `exerciseApi.saveExercises(toSync)`.
- On success (API returns 201 and result entries with `status: 'saved'`) it:
  - dispatches `MARK_SYNCED` for the saved ids (sets `syncStatus: 'synced'`).
  - removes those ids from the pending list via `exerciseStorage.deletePendingExercises(ids)`.
  - updates the local saved cache (`exercises:saved`) to mark the synced items.
- On failure it dispatches `MARK_SYNC_FAILED` to record failed sync ids (tracked in `failedSyncIds` set).

Local user actions and how they are handled:

- `saveExercise(name)` — creates a `SavedExercise` with a generated id (`Date.now() + random`), timestamps, `syncStatus: 'pending'`; dispatches `ADD_EXERCISE` and writes to pending storage (`savePendingExercise`). Returns the created `SavedExercise`.
- `updateExercise(exercise)` — updates `updatedAt`, sets `syncStatus: 'pending'`, dispatches `UPDATE_EXERCISE`, and saves to pending storage.
- `archiveExercise(savedExerciseId)` — dispatches `ARCHIVE_EXERCISE` (sets `archivedAt`), updates local saved cache entry to `syncStatus: 'synced'`, and calls API `archiveExercise(.., true)` asynchronously.
- `restoreExercise(savedExerciseId)` — dispatches `RESTORE_EXERCISE`, updates local saved cache to `syncStatus: 'synced'`, calls API `archiveExercise(.., false)` asynchronously.

The context exposes convenience helpers used by UI components:

- `savedExercises` — active (non-archived) saved exercises.
- `allSavedExercises` — includes archived exercises.
- `getById(savedExerciseId)` — lookup helper.
- `syncExercises()` — manually trigger a sync of pending exercises.

## Where exercises are used in the UI

Examples of files that read/update exercises:

- `packages/app/src/screens/SavedExercisesScreen.tsx` — list and management of saved exercises.
- `packages/app/src/components/ExerciseEditorModal.tsx` — edit a saved exercise (note/tags).
- `packages/app/src/components/ExerciseCard.tsx` — small UI component to render an exercise entry.
- `packages/app/src/components/AddExercise.tsx` — quick-add UI used when adding exercises to workouts.

## Example: create a SavedExercise

```ts
const now = new Date().toISOString();
const newExercise: SavedExercise = {
  savedExerciseId: `${Date.now()}_${Math.random().toString(36).slice(2,9)}`,
  name: 'Barbell Back Squat',
  note: '',
  tags: ['squat','compound','legs'],
  archivedAt: null,
  createdAt: now,
  updatedAt: now,
  syncStatus: 'pending',
};
```

## Notes & recommendations

- The client intentionally separates `pending` and `saved` caches so UI updates are immediate and resilient to network failures.
- `syncStatus` is the single field used to determine if an exercise still needs to be sent to the server.
- `failedSyncIds` is kept in-memory (as `Set<string>`) to avoid repeatedly trying to sync items that already failed until they are changed.

If you want additional documentation (examples of mutation flows, sequence diagrams, or API contract examples), tell me what format you prefer and I will add it.
