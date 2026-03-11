
# Workouts.md

This file contains the information relating to the workouts and their data structure and uses

## Workout Data

```SQL
CREATE TABLE workouts (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL references users(id),
    status STRING NOT NULL,
    start TIMESTAMP NOT NULL,
    end TIMESTAMP
    );
```

## Feature

- Users will be able to "start a workout" which will create a new workout.
- Users will then be able to add exercises to the workout, each exercise will be its own entry in the exercises table and will be linked to the workout via the workout_id field.

## Implementation

### AC

1. When I open the App
2. I can start a workout
    a. I can't start a second workout when I already have an active workout
    b. The active workout is hightlighted in the UI on the homepage to give easy access to the active workout
3. Then I can end the workout
    a. I will see the workout was created / saved in the database.

