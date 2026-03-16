import { getCurrentSession } from './auth';
import type { Workout } from '../types/workout';

export interface SaveWorkoutsResult {
  success: boolean;
  status: number;
  results: {
    workoutId: string;
    status: 'saved' | 'error';
  }[];
}

const baseUrl = process.env.EXPO_PUBLIC_API_DOMAIN;

export async function saveWorkouts(
  workouts: Workout | Workout[],
): Promise<SaveWorkoutsResult> {
  try {
    const session = await getCurrentSession();
    if (!session) {
      return { success: false, status: 401, results: [] };
    }

    const response = await fetch(`${baseUrl}/save`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.tokens.idToken}`,
      },
      body: JSON.stringify(workouts),
    });

    const responseData = await response.json();
    const results = responseData.results;

    return {
      success: response.status === 201,
      status: response.status,
      results,
    };
  } catch {
    return { success: false, status: 0, results: [] };
  }
}

export const getWorkouts = async (lastSK: string | undefined): Promise<{
	nextSk: string,
	workouts: Workout[],
}> => {
	try {
		const session = await getCurrentSession();
		const response = await fetch(`${baseUrl}/read${lastSK ? `?sk=${lastSK}` : ""}`, {
			headers: {
				Authorization: `Bearer ${session?.tokens.idToken}`
			},
		});
		const result = await response.json();
		return result;
	} catch (error) {
		console.error(error);
		return {
			nextSk: "",
			workouts: []
		};
	}
}

