# F0001 - Saved Workouts Grace Period

This file will outline the feature to allow users a "grace" period when editing previously recorded exercises

## Referance
- "workout": is a data item that is comprised of information read more at [workout.md](../info/workout.md)
- "exercise": is a data item that is a part of a workout read more at [exercise.md](../info/exercise.md)
- "target workout": is the workout the user wishes to update or restore to active
- "target exercise": is the exercise the user wishes to update
- describing an workout as "saved": the act of completing the save procedure for an workout read more at [save-procedure.md](../info/save-procedure.md)
- "restore": return the target workout to the active state
- "update": update the notes of either the target workout or target exercise
- "active workout": is the workout that is currently active when the user initiates a "restore" event

## Notes

- the target workout is "saved" at the "completedAt" time, which is updated if the user restores the exercise to the active state, and saves the exercise again.
    - this is the time used to calculate the grace period
- Each update functionality will have it's own lambda function registered but can be generic with the use of environment variables to edit what the cutoff time and the target value is.


### Restoring a workout to an active state
- timeframe: up to 2 hours after workout completion

#### Behaviour
- restores the exercise to the active status
- if there is already an active exercise warn the user that any unsaved exercise progress will be lost.
    - ask if they want to discard or save the currently active workout, with an option to cancel the operation
        - if the user choses the save option - follow the procedure for saving an exercise
        - if the user choses the discard option - destructively write the target exercise over the currently active exercise
- Upon saving the exercise follow the procedure for saving an exercise

### Updating body weight data
- timeframe: up to 6 hours after a workout completion

#### Behaviour
- The user would be presented with a button to "update" or "add" their body weight to the workout data, at the bottom of the page
- The frontend code will validate if the button is to be displayed or not
    - frontend will use the 6 hour cut off time
- The lambda function will further validate that the user is allowed to update the BW data
    - backend will use a 7 hour cut off time to allow for on the edge of time updates

### Updating workout notes and exercise data
- timeframe: up to 2 days after a workout is completed

#### Behaviour
- The user will be presented with a button to "edit" or "add" to the already existing note, underneath the text within the note box
- The user will be presented with a button to "edit" or "add" to the already existing note, underneath the text within the note box for an exercise

- The frontend code will validate if the option is available
    - the frontend will use 48 hours as the cut off time
- The lambda function will further validate that the user is allowed to update the note data
    - the lambda will use 49 hours as the cut off time


