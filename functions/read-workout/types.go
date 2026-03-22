package main

// WorkoutSet represents a single set within an exercise.
type WorkoutSet struct {
	SetID      string   `json:"setId" dynamodbav:"setId"`
	Order      int      `json:"order" dynamodbav:"order"`
	Reps       *int     `json:"reps" dynamodbav:"reps"`
	Weight     *float64 `json:"weight" dynamodbav:"weight"`
	WeightUnit string   `json:"weightUnit" dynamodbav:"weightUnit"`
}

// WorkoutExercise represents an exercise performed during a workout.
type WorkoutExercise struct {
	ExerciseID      string       `json:"exerciseId" dynamodbav:"exerciseId"`
	SavedExerciseID *string      `json:"savedExerciseId" dynamodbav:"savedExerciseId"`
	Name            string       `json:"name" dynamodbav:"name"`
	Order           int          `json:"order" dynamodbav:"order"`
	Sets            []WorkoutSet `json:"sets" dynamodbav:"sets"`
}

// Workout is the top-level item stored in DynamoDB.
type Workout struct {
	// Partition key — user identifier.
	UserID string `json:"userId,omitempty" dynamodbav:"userId"`
	// Sort key — WORKOUT#<timestamp>#<workoutId>
	SK          string            `json:"sk,omitempty" dynamodbav:"sk"`
	WorkoutID   string            `json:"workoutId,omitempty" dynamodbav:"workoutId"`
	TemplateID  *string           `json:"templateId" dynamodbav:"templateId"`
	StartedAt   string            `json:"startedAt" dynamodbav:"startedAt"`
	CompletedAt *string           `json:"completedAt" dynamodbav:"completedAt"`
	Notes       string            `json:"notes" dynamodbav:"notes"`
	Tags        []string          `json:"tags" dynamodbav:"tags"`
	BodyWeight  *float32          `json:"bodyWeight" dynamodbav:"bodyWeight"`
	Exercises   []WorkoutExercise `json:"exercises" dynamodbav:"exercises"`
	CreatedAt   string            `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   string            `json:"updatedAt" dynamodbav:"updatedAt"`
}
