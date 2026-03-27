Feature: Exercise Instance Notes

Overview
--------
Users need the ability to attach a single, per-instance note to an exercise entry inside a workout. This is not the same as an exercise-level perpetual note (which applies to the exercise type across workouts). Instead, the note is a one-off text field that belongs only to that occurrence of the exercise in a specific workout session.

User story
----------
As a user, when I add or edit an exercise inside a workout, I want to add a short note that applies only to that exercise instance (for example: "beltless today", "felt weak on last set", "paused at 2s"). I expect to see that note alongside the exercise when viewing the workout and be able to edit or remove it.

Goals
-----
- Provide a lightweight, editable note attached to a single exercise instance in a workout.
- Keep the note distinct from the exercise template/permanent notes.
- Allow creating, updating, displaying, and deleting the note from the workout UI and API.

Acceptance criteria
-------------------
- Users can add a text note (optional) when creating or editing an exercise within a workout.
- The note is stored with the exercise instance only and is returned with workout/exercise instance data.
- The note can be edited or cleared after creation.
- Existing exercise-level/perpetual notes remain unchanged and separate.
- UI shows the note in workout detail and in any compact view where exercise instance details are visible.

Data model
----------
- Add a nullable `note` (string) field to the exercise-instance model/entity (e.g. `workout_exercises`, `exercise_entries`, or equivalent). Keep length reasonable (e.g. 1000 chars) and trim whitespace.
- No changes to the exercise template/entity.

API
---
- Update workout create/update endpoints that accept exercise instances to accept an optional `note` property on each exercise item.
- Ensure GET workout endpoints include `note` on exercise instances in the response.
- Add validation: `note` is optional, max length enforced, sanitize/escape output to prevent XSS.

UI / UX
-------
- Keep the exercise row compact by showing an `Add note` button (or a small note icon) instead of an always-visible input field.
- When the user taps/clicks `Add note`, create and reveal a text input area (single-line or small textarea) in-place and automatically focus it so the user can start typing immediately.
- If the user blurs/unfocuses the input without entering any valid text (empty or whitespace-only), remove the input area and return to the initial `Add note` button state to avoid clutter.
- When a note exists, display it unobtrusively under the exercise title or as an expandable inline element; allow tapping/clicking the note or an edit icon to reopen the input for editing.
- Include a clear/remove action to delete the note; on mobile, keep the interaction compact and make the input collapsible.

Security & validation
---------------------
- Enforce server-side max length and sanitize input to avoid stored XSS.
- Rate-limit or size-limit per-request to avoid abuse.

Migration / Backwards compatibility
----------------------------------
- Add the `note` column to the exercise-instance table as nullable with default NULL; no migration of existing data required.
- API changes are additive (optional field), so older clients continue working.

Testing
-------
- Unit tests for serializer/model to persist `note` values and enforce length/sanitization.
- API tests for create/update/get including notes.
- UI tests (integration) to confirm adding, editing, displaying, and deleting notes across viewport sizes.

Rollout
-------
1. Add DB column and API + backend validation behind a feature flag if desired.
2. Add UI controls to exercise edit/create flows and display in workout detail.
3. Monitor for errors and gather user feedback.

Edge cases
----------
- Empty or whitespace-only notes are treated as "no note" and stored/returned as null.
- Very long notes are rejected with a clear validation message.
- When copying or duplicating workouts, decide whether to copy instance notes (recommended: copy only when explicitly duplicating full workout content).

Open questions
--------------
- Should exercise instance notes be searchable across workouts? (If yes, index strategy needed.)
- When exporting a workout, should instance notes be included? (Recommended: yes.)
