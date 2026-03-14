import { getCurrentSession } from './auth';
import type { Workout } from '../types/workout';

export interface SaveWorkoutsResult {
  success: boolean;
  status: number;
}

export async function saveWorkouts(
  workouts: Workout | Workout[],
): Promise<SaveWorkoutsResult> {
  try {
    const session = await getCurrentSession();
    if (!session) {
      return { success: false, status: 401 };
    }

    const response = await fetch('/api/save', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.tokens.idToken}`,
      },
      body: JSON.stringify(workouts),
    });

    return {
      success: response.status === 200,
      status: response.status,
    };
  } catch {
    return { success: false, status: 0 };
  }
}
