# Chargeback Prevention Strategy

## 1. Statement Descriptor

Set a recognisable statement descriptor in the Stripe dashboard so users can identify the charge on their bank statement.

- **Stripe Dashboard > Settings > Public details**
- Set descriptor to something clear, e.g. `WORKOUT TRACKER` or `WRKOUT TRKR`
- Max 22 characters, uppercase, no special characters
- Also set the shortened descriptor for card networks that truncate

Prevents "I don't recognise this charge" disputes, which are the most common chargeback reason.

## 2. Receipt and Renewal Emails

Enable automatic Stripe emails so users are reminded what they're paying for.

- **Stripe Dashboard > Settings > Emails**
- Enable **Successful payment** receipts — sent after each charge
- Enable **Upcoming renewal** reminders — sent before the next billing cycle, giving users a chance to cancel rather than dispute

## 3. Easy Self-Service Cancellation

Users who can't figure out how to cancel will call their bank instead, resulting in a chargeback.

- Stripe Customer Portal handles cancellation, payment method updates, and invoice history
- Surface the "Manage Subscription" button prominently in the app settings — not buried in sub-menus
- Use **cancel-at-end-of-period** rather than immediate cancellation — users keep access until the period ends, reducing "I cancelled but was still charged" confusion

## 4. Customer Email on Checkout

When creating a Stripe Checkout Session, include the user's email (from their Cognito profile) via `customer_email` or on the Stripe Customer object. This ensures:

- Stripe can send receipts to the correct address
- Contact details are available if a dispute arises

## 5. Refund Before It Escalates

A refund costs nothing beyond the original transaction amount. A chargeback costs the transaction amount plus a dispute fee (~£15–20), and counts against the dispute rate.

- Provide a visible support email or contact method
- If someone requests a refund, issue it — it is always cheaper than a chargeback
- Stripe supports programmatic refunds if a self-service refund flow is needed later

## 6. Dispute Response Preparation

When a chargeback occurs, Stripe provides a window to submit evidence. Have these ready:

- **Proof of service**: DynamoDB logs showing the user actively used the app (workout records, login timestamps)
- **Terms of service / refund policy**: Published on the website — even a single paragraph helps
- **Communication records**: Any emails or support interactions with the user

Stripe's dispute dashboard guides through what evidence to submit for each dispute reason code.

## 7. Stripe Radar (Fraud Detection)

Stripe Radar is included with every Stripe account and runs automatically on every payment.

- Blocks known fraudulent cards
- Flags suspicious transactions (mismatched billing address, velocity checks, etc.)
- No configuration needed — enabled by default
- Low risk for a £4/month subscription (fraudsters prefer high-value single charges), but Radar catches card-testing if it happens

## 8. Avoid Free Trials with Card Upfront

Free trials requiring a card upfront have higher chargeback rates because users forget they signed up. The freemium model (30 free workouts, no card required) avoids this entirely. When users enter their card, it is a deliberate decision to pay.

## Priority Order

| Priority | Action | Effort |
|----------|--------|--------|
| 1 | Set clear statement descriptor | 2 minutes in Stripe dashboard |
| 2 | Enable receipt and renewal emails | 1 minute in Stripe settings |
| 3 | Easy cancellation via Stripe Customer Portal | Already implemented — ensure it's discoverable in the UI |
| 4 | Publish a refund policy page | Short page on the domain |
| 5 | Include customer email on Checkout | Code change in checkout-create-session Lambda |
| 6 | Prepare dispute evidence workflow | Document internal process |
| 7 | Stripe Radar | No action needed — on by default |
| 8 | No card-upfront free trials | No action needed — freemium model avoids this |
