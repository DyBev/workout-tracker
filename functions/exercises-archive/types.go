package main

// SavedExercise represents a saved exercise item in DynamoDB.
type SavedExercise struct {
	UserID          string   `json:"userId,omitempty" dynamodbav:"userId"`
	SK              string   `json:"sk,omitempty" dynamodbav:"sk"`
	SavedExerciseID string   `json:"savedExerciseId" dynamodbav:"savedExerciseId"`
	Name            string   `json:"name" dynamodbav:"name"`
	Note            string   `json:"note" dynamodbav:"note"`
	Tags            []string `json:"tags" dynamodbav:"tags"`
	ArchivedAt      *string  `json:"archivedAt" dynamodbav:"archivedAt"`
	CreatedAt       string   `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt       string   `json:"updatedAt" dynamodbav:"updatedAt"`
}

// ArchiveRequest is the expected JSON body for the archive endpoint.
type ArchiveRequest struct {
	SavedExerciseID string `json:"savedExerciseId"`
	Archive         bool   `json:"archive"`
}
