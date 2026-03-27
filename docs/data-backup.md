# Data Backup and Export — Implementation Notes

This document outlines a simple, user‑focused data backup/export feature for Workout Tracker. The goal is to let users download their workouts and exercises in a portable format and to provide guidance for implementing server-side and client-side flows.

Goals
- Give users a clear, one-click way to export their data (workouts, exercises, templates, account metadata).
- Provide a machine-readable export (JSON) and a human-readable export (CSV for workouts and exercises).
- Respect user privacy and security: exports require authentication and are delivered securely.

User flows
1. On-demand export (recommended)
   - User navigates to Account > Settings > Data & privacy > Export my data.
   - User clicks `Export my data` and confirms via a modal.
   - The system creates an export bundle (JSON + CSVs), stores it temporarily, and notifies the user when ready. The user downloads it from a secure URL that expires after N hours (recommended: 24h).

2. Immediate small export
   - For small accounts, the export can be generated synchronously and returned immediately as a file download.

3. Email delivery (optional)
   - Optionally, allow the user to request the export be emailed to their verified account address.

Data included
- Account metadata: creation date.
- Workouts: date, duration, exercises (id, name, sets, reps, weight, notes).
- Exercises: id, name, notes, tags.
- Templates: name, structure.

Formats
- Primary: single JSON file (export.json) that contains full records and relationships.
- Secondary: CSVs for `workouts.csv` and `exercises.csv` for easy spreadsheet import.

Security
- Require the user to be authenticated to request an export.
- Generate export files in a private storage area with randomized filenames and short expiry (e.g., pre-signed S3 URL expiring in 24 hours).
- If emailing exports, attach the file or provide the secure download link. Verify the user's email is confirmed before delivering.
- Rate limit exports per account (e.g., 1 per week) to avoid abuse. _users can contact support@dybev.uk if they exceed / require exceptions, this information will be in the terms_

Implementation notes (server-side)
- Endpoint: `POST /api/v1/account/export` — creates an export job, returns job id and estimated completion. Creates the jobID entry in the dynamoDB under the userID PK, with a TTL set to 7 days.
- Job processing(step function): background worker collects data, writes JSON/CSV to temporary storage, generates a short-lived download link, marks job as complete, saves the data to dynamoDB and sets the TTL for the S3 URLs to 24h, the client will request new signed URLs with a confirmation step sent to their emails.
- Endpoint: `GET /api/v1/account/export/:jobId` — returns job status and, when ready, the download URL(s).
- Endpoint: `GET /api/account/export/:jobId/urls` - returns and sets the new signed URLs for the exported data with a 24h TTL on dynamoDB, only valid and equal user IDs can request this, confirmation is sent to the email of the account holder.
- Storage: use cloud object storage (S3) with lifecycle policy to expire files after X days (recommended 7 days).
- Permissions: stored files must not be publicly accessible; use pre-signed URLs.

Implementation notes (client-side)
- Add a Settings UI with an Export button and a small explanation of included data and download expiry.
- Add a check status button with a 15 minute cooldown to prevent spamming.
- Provide direct download buttons for JSON and CSV files when ready.
- Show a one-time confirmation and a warning about sensitive data in exported files when requesting the export.

Testing
- Unit test serializers to ensure JSON contains expected fields and relationships.
- Integration test the end-to-end export job (create sample data, request export, verify files created and contents match).
- Security tests: ensure unauthenticated requests are rejected and pre-signed URLs expire.

Operational considerations
- Monitor storage usage and apply lifecycle rules.
- Monitor export job failures and alert on unusual error rates.
- Consider adding export throttling if abuse is detected. Doubling the period if the user makes the maximum requests per month. Can save the period in the user metadata entry in dynamoDB.

If you want, I can produce API request/response examples and a sample JSON schema for `export.json` next.
