import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import * as subscriptionApi from '../services/subscriptionApi';
import * as subscriptionStorage from '../services/subscriptionStorage';
import type { SubscriptionStatus } from '../services/subscriptionApi';

export interface SubscriptionState {
  isLoading: boolean;
  status: SubscriptionStatus['status'];
  subscriptionSource?: SubscriptionStatus['subscriptionSource'];
  currentPeriodEnd?: number;
}

type SubscriptionAction =
  | { type: 'SET_LOADING'; isLoading: boolean }
  | { type: 'SET_STATUS'; payload: SubscriptionStatus };

const initialState: SubscriptionState = {
  isLoading: true,
  status: 'none',
};

function subscriptionReducer(
  state: SubscriptionState,
  action: SubscriptionAction,
): SubscriptionState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.isLoading };
    case 'SET_STATUS':
      return {
        isLoading: false,
        status: action.payload.status,
        subscriptionSource: action.payload.subscriptionSource,
        currentPeriodEnd: action.payload.currentPeriodEnd,
      };
    default:
      return state;
  }
}

export interface SubscriptionContextValue {
  state: SubscriptionState;
  isSubscribed: boolean;
  refreshStatus: () => Promise<void>;
}

const SubscriptionContext = createContext<SubscriptionContextValue | null>(null);

interface SubscriptionProviderProps {
  children: React.ReactNode;
}

export function SubscriptionProvider({ children }: SubscriptionProviderProps) {
  const [state, dispatch] = useReducer(subscriptionReducer, initialState);

  const refreshStatus = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', isLoading: true });
    try {
      const result = await subscriptionApi.getSubscriptionStatus();
      dispatch({ type: 'SET_STATUS', payload: result });
      await subscriptionStorage.setCachedSubscriptionStatus(result);
    } catch {
      // API unavailable (offline) — fall back to cached status
      try {
        const cached = await subscriptionStorage.getCachedSubscriptionStatus();
        if (cached) {
          dispatch({ type: 'SET_STATUS', payload: cached });
          return;
        }
      } catch {
        // Cache read failed — fall through to default
      }
      dispatch({
        type: 'SET_STATUS',
        payload: { status: 'none' },
      });
    }
  }, []);

  useEffect(() => {
    refreshStatus();
  }, [refreshStatus]);

  const isSubscribed = state.status === 'active';

  const value = useMemo<SubscriptionContextValue>(
    () => ({
      state,
      isSubscribed,
      refreshStatus,
    }),
    [state, isSubscribed, refreshStatus],
  );

  return (
    <SubscriptionContext.Provider value={value}>
      {children}
    </SubscriptionContext.Provider>
  );
}

export function useSubscription(): SubscriptionContextValue {
  const context = useContext(SubscriptionContext);
  if (!context) {
    throw new Error('useSubscription must be used within a SubscriptionProvider');
  }
  return context;
}
