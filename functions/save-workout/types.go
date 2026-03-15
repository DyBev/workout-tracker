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
	ExerciseID string       `json:"exerciseId" dynamodbav:"exerciseId"`
	Name       string       `json:"name" dynamodbav:"name"`
	Order      int          `json:"order" dynamodbav:"order"`
	Sets       []WorkoutSet `json:"sets" dynamodbav:"sets"`
}

// BodyWeight records the user's body weight at the time of the workout.
type BodyWeight struct {
	Value float64 `json:"value" dynamodbav:"value"`
	Unit  string  `json:"unit" dynamodbav:"unit"`
}

// Workout is the top-level item stored in DynamoDB.
// It mirrors the TypeScript Workout interface from the frontend.
type Workout struct {
	// Partition key — user identifier.
	UserID string `json:"userId" dynamodbav:"userId"`
	// Sort key — WORKOUT#<timestamp>#<workoutId>.
	SK          string            `json:"sk" dynamodbav:"sk"`
	WorkoutID   string            `json:"workoutId" dynamodbav:"workoutId"`
	TemplateID  *string           `json:"templateId" dynamodbav:"templateId"`
	StartedAt   string            `json:"startedAt" dynamodbav:"startedAt"`
	CompletedAt *string           `json:"completedAt" dynamodbav:"completedAt"`
	Notes       string            `json:"notes" dynamodbav:"notes"`
	Tags        []string          `json:"tags" dynamodbav:"tags"`
	BodyWeight  *BodyWeight       `json:"bodyWeight" dynamodbav:"bodyWeight"`
	Exercises   []WorkoutExercise `json:"exercises" dynamodbav:"exercises"`
	CreatedAt   string            `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   string            `json:"updatedAt" dynamodbav:"updatedAt"`
}

type SaveResult struct {
	WorkoutID string `json:"workoutId"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}
