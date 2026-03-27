package main

// UserProfile represents a user profile item in the UserProfiles DynamoDB table.
type UserProfile struct {
	UserID             string  `json:"userId" dynamodbav:"userId"`
	StripeCustomerID   *string `json:"stripeCustomerId,omitempty" dynamodbav:"stripeCustomerId,omitempty"`
	SubscriptionID     *string `json:"subscriptionId,omitempty" dynamodbav:"subscriptionId,omitempty"`
	SubscriptionStatus string  `json:"subscriptionStatus" dynamodbav:"subscriptionStatus"`
	SubscriptionSource string  `json:"subscriptionSource" dynamodbav:"subscriptionSource"`
	CurrentPeriodEnd   *int64  `json:"currentPeriodEnd,omitempty" dynamodbav:"currentPeriodEnd,omitempty"`
	CreatedAt          string  `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt          string  `json:"updatedAt" dynamodbav:"updatedAt"`
}

// StatusResponse is the JSON response returned to the frontend.
type StatusResponse struct {
	Status             string `json:"status"`
	SubscriptionSource string `json:"subscriptionSource,omitempty"`
	CurrentPeriodEnd   *int64 `json:"currentPeriodEnd,omitempty"`
}
