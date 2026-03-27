import { getCurrentSession } from './auth';

const baseUrl = process.env.EXPO_PUBLIC_API_DOMAIN;

export interface SubscriptionStatus {
  status: 'active' | 'past_due' | 'cancelled' | 'none';
  subscriptionSource?: 'manual' | 'stripe';
  currentPeriodEnd?: number;
}

/**
 * Fetches the subscription status from the API.
 *
 * Returns the status on success, or `{ status: 'none' }` when the user has
 * no session or the server returns a non-OK response with a parseable body.
 *
 * Throws on network errors so the caller can fall back to a cached value
 * when the device is offline.
 */
export async function getSubscriptionStatus(): Promise<SubscriptionStatus> {
  const session = await getCurrentSession();
  if (!session) {
    return { status: 'none' };
  }

  const response = await fetch(`${baseUrl}/subscription/status`, {
    headers: {
      Authorization: `Bearer ${session.tokens.idToken}`,
    },
  });

  if (!response.ok) {
    return { status: 'none' };
  }

  const result = await response.json();
  return {
    status: result.status ?? 'none',
    subscriptionSource: result.subscriptionSource,
    currentPeriodEnd: result.currentPeriodEnd,
  };
}
