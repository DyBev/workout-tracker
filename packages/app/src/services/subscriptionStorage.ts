import AsyncStorage from '@react-native-async-storage/async-storage';
import type { SubscriptionStatus } from './subscriptionApi';

const SUBSCRIPTION_STATUS_KEY = 'subscription:status';

export async function getCachedSubscriptionStatus(): Promise<SubscriptionStatus | null> {
  const raw = await AsyncStorage.getItem(SUBSCRIPTION_STATUS_KEY);
  if (!raw) return null;
  return JSON.parse(raw) as SubscriptionStatus;
}

export async function setCachedSubscriptionStatus(status: SubscriptionStatus): Promise<void> {
  await AsyncStorage.setItem(SUBSCRIPTION_STATUS_KEY, JSON.stringify(status));
}

export async function clearCachedSubscriptionStatus(): Promise<void> {
  await AsyncStorage.removeItem(SUBSCRIPTION_STATUS_KEY);
}
