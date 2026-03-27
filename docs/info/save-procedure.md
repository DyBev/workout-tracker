# Save procedure

This document describes what happens during a "save" event in Workout Tracker. It is written to be human-readable and to provide structured context for LLMs (clear headings, explicit data shapes, and code references).

## Summary
- Purpose: explain the local persistence and server sync flow for workouts and saved exercises.
- Key modules: `packages/app/src/services/*` and `packages/app/src/contexts/*` (see references below).
- Principle: offline-first — writes go to local storage first, then the app attempts to sync to the server when possible.

## High-level flow
1. User action creates or updates an object (workout or saved exercise) in UI.
2. Context updates in-memory state with `syncStatus: 'pending'` and timestamps.
3. The item is persisted to AsyncStorage (a pending queue or active key).
4. The app attempts to send the item(s) to the server via API helpers.
5. On API success the item is marked `synced` locally and removed from the pending queue; on failure it remains pending and is retried later.

## Workouts

### Where the code lives
- Local storage helpers: `packages/app/src/services/workoutStorage.ts`
- API client: `packages/app/src/services/workoutApi.ts`
- Context and orchestration: `packages/app/src/contexts/WorkoutContext.tsx`

### Local persistence
- Active workout: AsyncStorage key `workout:active` (helpers: `saveActiveWorkout`, `getActiveWorkout`, `clearActiveWorkout`).
- Pending workouts queue: AsyncStorage key `workout:pending` (helpers: `savePendingWorkout`, `getPendingWorkouts`, `deletePendingWorkouts`).

### Completing a workout (actual steps)
1. `completeWorkout` (in `WorkoutContext`) builds a `completed` Workout object with `syncStatus: 'pending'` and sets `completedAt` / `updatedAt`.
2. `workoutStorage.savePendingWorkout(completed)` adds it to `workout:pending` (prevents duplicates by `workoutId`).
3. `workoutStorage.clearActiveWorkout()` removes the active key.
4. Local state is updated (`COMPLETE_WORKOUT`) so the UI reflects the change.
5. `syncWorkouts([completed])` is called asynchronously to try uploading.

### Sync to server
- Function: `workoutApi.saveWorkouts(workouts)`
- Endpoint: `POST ${EXPO_PUBLIC_API_DOMAIN}/save` (sends a workout or array of workouts)
- Authentication: uses `getCurrentSession()` and sends `Authorization: Bearer <idToken>`

Response and handling
- `SaveWorkoutsResult` (client):

```json
{
  "success": true | false,
  "status": 201 | <other>,
  "results": [
    { "workoutId": "<id>", "status": "saved" | "error" }
  ]
}
```

- If `success === true` (HTTP 201):
  - context dispatches `MARK_SYNCED` for saved IDs,
  - `workoutStorage.deletePendingWorkouts(ids)` removes them from `workout:pending`.
- If not successful: successful IDs in the response are processed as saved; failed IDs are left pending and added to `failedSyncIds` for retry logic.

### Startup behavior
- On app restore (`WorkoutProvider` restore flow) the provider reads: active workout, pending workouts, and server workouts (`workoutApi.getWorkouts`).
- It merges server and pending sets, restores an active workout if appropriate, and calls `syncWorkouts(pending)` if pending items exist.

## Saved exercises

### Where the code lives
- Local storage helpers: `packages/app/src/services/exerciseStorage.ts`
- API client: `packages/app/src/services/exerciseApi.ts`
- Context and orchestration: `packages/app/src/contexts/ExerciseContext.tsx`

### Local persistence
- Saved (cached) exercises list: AsyncStorage key `exercises:saved` (`getSavedExercises`, `setSavedExercises`, `updateExercise`).
- Pending saved-exercises queue: AsyncStorage key `exercises:pending` (`savePendingExercise`, `getPendingExercises`, `deletePendingExercises`).

### Save / update flow
1. Creating or updating a saved exercise updates in-memory state with `syncStatus: 'pending'`.
2. `exerciseStorage.savePendingExercise(exercise)` writes it to `exercises:pending` (inserts or updates by `savedExerciseId`).
3. `syncExercises()` is used to upload pending or unsynced exercises.

### Sync to server
- Function: `exerciseApi.saveExercises(exercises)`
- Endpoint: `POST ${EXPO_PUBLIC_API_DOMAIN}/exercises/save` (sends an array)
- Authentication: uses `getCurrentSession()` and `Authorization: Bearer <idToken>`

Response and handling
- `SaveExercisesResult` (client):

```json
{
  "success": true | false,
  "status": 201 | <other>,
  "results": [
    { "savedExerciseId": "<id>", "status": "saved" | "error" }
  ]
}
```

- If `success === true` (HTTP 201):
  - context dispatches `MARK_SYNCED` for saved IDs,
  - `exerciseStorage.deletePendingExercises(ids)` removes them from `exercises:pending`,
  - cached saved exercises (`exercises:saved`) are updated to reflect `syncStatus: 'synced'` via `exerciseStorage.setSavedExercises`.
- If not successful: successful IDs are processed; failed IDs remain pending and are added to `failedSyncIds`.

### Startup behavior
- On provider restore the app reads `exercises:pending` and the server list (`exerciseApi.readExercises()`), merges them and updates the local cache, then calls `syncExercises(pending)` if there are pending items.

## Authentication and API gating
- All server calls call `getCurrentSession()` (`packages/app/src/services/auth.ts`). If no valid session is available the API helpers return early (usually with `success: false` or empty result) and no sync occurs.
- The client sends the Cognito ID token in `Authorization: Bearer <idToken>` headers.

## Subscription & limits (terms reference)
- See `docs/terms.md` for policy: free tier allows up to 60 cloud-synced workouts and 10 saved exercises.
- Behaviour per terms:
  - If a user exceeds free-tier cloud limits, the app keeps working locally but will not sync additional items to cloud storage until subscription is active.
  - When a user subscribes, locally stored items that were prevented from syncing should be uploaded on next sign-in (the app already attempts pending syncs on provider restore/startup).

## Error handling and retry strategy
- API helpers catch network errors and return a failure result (`status: 0` or `success: false`). Contexts trap exceptions and keep pending queues intact.
- Partial success: successful IDs are removed/marked synced; failures remain pending.
- Retry triggers:
  - automatic on app startup / provider restore if pending items exist,
  - automatic after completing items (e.g., `completeWorkout` calls `syncWorkouts([completed])`),
  - manual re-sync can be triggered by calling `syncWorkouts()` or `syncExercises()` from the contexts.

## AsyncStorage keys (quick reference)
- `workout:active` — active workout (single object)
- `workout:pending` — array of pending workouts
- `exercises:saved` — cached saved exercises
- `exercises:pending` — array of pending saved exercises
- `subscription:status` — cached subscription status (`packages/app/src/services/subscriptionStorage.ts`)

## API endpoints (client-side references)
- Workouts:
  - POST `${EXPO_PUBLIC_API_DOMAIN}/save` — save workouts
  - GET `${EXPO_PUBLIC_API_DOMAIN}/read?sk=<sk>` — read workouts list (`workoutApi.getWorkouts`)
- Exercises:
  - POST `${EXPO_PUBLIC_API_DOMAIN}/exercises/save` — save exercises
  - GET `${EXPO_PUBLIC_API_DOMAIN}/exercises/read` — read exercises
  - PATCH `${EXPO_PUBLIC_API_DOMAIN}/exercises/archive` — archive/restore
- Subscription:
  - GET `${EXPO_PUBLIC_API_DOMAIN}/subscription/status`

## Useful code locations (quick links)
- `packages/app/src/contexts/WorkoutContext.tsx` — orchestration for workouts (start, update, complete, sync)
- `packages/app/src/services/workoutStorage.ts` — AsyncStorage helpers for workouts
- `packages/app/src/services/workoutApi.ts` — workout API client
- `packages/app/src/contexts/ExerciseContext.tsx` — orchestration for saved exercises
- `packages/app/src/services/exerciseStorage.ts` — AsyncStorage helpers for exercises
- `packages/app/src/services/exerciseApi.ts` — exercise API client
- `packages/app/src/services/auth.ts` — Cognito session helpers used by API calls
- `docs/terms.md` — subscription limits and relevant policy notes

## Examples (pseudo sequence)
- Complete a workout while offline:
  1. UI calls `completeWorkout()` → writes to `workout:pending` and clears `workout:active`.
  2. `syncWorkouts([completed])` attempts network save and fails; workout remains in `workout:pending`.
  3. On next app start, provider restore sees pending items and calls `syncWorkouts(pending)` again.

- Save a new exercise while online:
  1. `saveExercise()` creates a `SavedExercise` with `syncStatus: 'pending'` and writes it to `exercises:pending`.
  2. `syncExercises()` posts to server; on success the exercise is marked `synced` and removed from `exercises:pending` and the saved cache is updated.

## Suggested improvements (optional)
1. Add explicit exponential backoff and scheduled retries for failed syncs.
2. Expose a visible sync status/notification in the UI to inform users of pending items and quota limits.
3. Add telemetry around partial failures for diagnostics.

---
Generated from code in `packages/app/src` and policy notes in `docs/terms.md`.
