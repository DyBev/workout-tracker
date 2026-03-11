
# Exercises.md

This file contains the information relating to the exercises and their data structure and uses

## Exercise Data

```SQL
CREATE TABLE exercises (
    id UUID PRIMARY KEY, 
    workout_id UUID NOT NULL,
    exercise_id INTEGER NOT NULL,
    );
```

- `id` is the id of the exercise instance which will be linked to the workout via the workout_id field and will be used to link the sets to the exercise via the exercise_id field
- `exercise_id` is the id of the exercise which will be used to link to the exercise metadata


```SQL
CREATE TABLE sets (
    id UUID PRIMARY KEY,
    exercise_id UUID NOT NULL,
    set_number INTEGER NOT NULL,
    reps INTEGER,
    value FLOAT,
    unit STRING,
    time INTEGER,
    );
```

- Each set will have it's own entry in the sets table and will be linked to the exercise via the exercise_id field

## Feature

- Users will be able to add exercises to the workout
- Users will be able to add sets to the exercises
- Users will be able to record weight and reps for each set, with reps, value, and unit being optional to allow for exercises that don't require reps, value, or units to be recorded
- Users will be able to recod the time if they wish, time will be stored as an integer in secnods
