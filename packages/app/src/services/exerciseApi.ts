import { getCurrentSession } from './auth';
import type { SavedExercise } from '../types/workout';

const baseUrl = process.env.EXPO_PUBLIC_API_DOMAIN;

export async function readExercises(): Promise<{ exercises: SavedExercise[] }> {
  try {
    const session = await getCurrentSession();
    if (!session) {
      return { exercises: [] };
    }

    const response = await fetch(`${baseUrl}/exercises/read`, {
      headers: {
        Authorization: `Bearer ${session.tokens.idToken}`,
      },
    });

    if (!response.ok) {
      return { exercises: [] };
    }

    const result = await response.json();
    return { exercises: result.exercises ?? [] };
  } catch {
    return { exercises: [] };
  }
}

export async function saveExercises(
  exercises: SavedExercise | SavedExercise[],
): Promise<{ success: boolean; status: number }> {
  try {
    const session = await getCurrentSession();
    if (!session) {
      return { success: false, status: 401 };
    }

    const response = await fetch(`${baseUrl}/exercises/save`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.tokens.idToken}`,
      },
      body: JSON.stringify(Array.isArray(exercises) ? exercises : [exercises]),
    });

    return {
      success: response.status === 201,
      status: response.status,
    };
  } catch {
    return { success: false, status: 0 };
  }
}

export async function archiveExercise(
  savedExerciseId: string,
  archive: boolean,
): Promise<{ success: boolean; status: number }> {
  try {
    const session = await getCurrentSession();
    if (!session) {
      return { success: false, status: 401 };
    }

    const response = await fetch(`${baseUrl}/exercises/archive`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.tokens.idToken}`,
      },
      body: JSON.stringify({ savedExerciseId, archive }),
    });

    return {
      success: response.ok,
      status: response.status,
    };
  } catch {
    return { success: false, status: 0 };
  }
}
