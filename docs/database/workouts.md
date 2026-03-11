
# Workouts.md

This file contains the information relating to the workouts and their data structure and uses

## Workout Data

```SQL
CREATE TABLE workouts (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    start DATE NOT NULL,
    );
```

## Feature

- Users will be able to "start a workout" which will create a new workout.
- Users will then be able to add exercises to the workout, each exercise will be it's own entry in the exercises table and will be linked to the workout via the workout_id field.

## Roadmap

### Initial Implementation

#### AC

1. When I open the App
2. I can start a workout
3. Then I can end the workout
    a. I will see the workout was created / saved in the database.

### Workout Templates - Improvement

#### Creating a workout template

##### Feature Definition

- Users will be able to creat a workout template which will consist of a name and a list of exercise UUIDs.
- The exercises won't be ordered to allow some flexibility in the workout

##### AC

1. When I open the App
2. I can create a workout template by adding exercises to the workout template and saving it
3. Then I can see the workout template in the list of workout templates
4. I can see the workout was saved in the database

#### Starting a workout from a saved workout template

##### Feature Definition

- Users will be able to start a workout from a saved workout template which will create a new workout with the exercises from the workout template
- Users will be able to select and update the exercises during the workout

##### AC

1. When I open the App
2. I can start a workout by selecting a saved workout template
3. Then I can end the workout
    a. I will see the workout was saved in the database.

#### Updating a workout template

##### Feature Definition

- Users will be able to update a workout template by editing the template via the template details page
- Updating a template will only updat the template and won't update any past workouts

##### AC

1. When I open the App
2. I can update a workout template by editing the template via the template details page
3. Then I can see the workout template was updated in the database
4. When I start a workout from the updated workout template I can see the updated exercises in the workout

