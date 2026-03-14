import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import * as authService from '../services/auth';
import type {
  AuthContextValue,
  AuthState,
  AuthUser,
  SignUpParams,
  ConfirmSignUpParams,
  SignInParams,
  ForgotPasswordParams,
  ResetPasswordParams,
  ChangePasswordParams,
} from '../types/auth';

type AuthAction =
  | { type: 'RESTORE_SESSION'; user: AuthUser }
  | { type: 'SIGN_IN'; user: AuthUser }
  | { type: 'SIGN_OUT' }
  | { type: 'SET_LOADING'; isLoading: boolean };

const initialState: AuthState = {
  user: null,
  isLoading: true,
  isAuthenticated: false,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'RESTORE_SESSION':
      return {
        user: action.user,
        isLoading: false,
        isAuthenticated: true,
      };
    case 'SIGN_IN':
      return {
        user: action.user,
        isLoading: false,
        isAuthenticated: true,
      };
    case 'SIGN_OUT':
      return {
        user: null,
        isLoading: false,
        isAuthenticated: false,
      };
    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.isLoading,
      };
    default:
      return state;
  }
}

const AuthContext = createContext<AuthContextValue | null>(null);

interface AuthProviderProps {
  children: React.ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  useEffect(() => {
    async function restoreSession() {
      try {
        const session = await authService.getCurrentSession();
        if (session) {
          dispatch({ type: 'RESTORE_SESSION', user: session.user });
        } else {
          dispatch({ type: 'SIGN_OUT' });
        }
      } catch {
        dispatch({ type: 'SIGN_OUT' });
      }
    }

    restoreSession();
  }, []);

  const signUp = useCallback(async (params: SignUpParams) => {
    await authService.signUp(params);
  }, []);

  const confirmSignUp = useCallback(async (params: ConfirmSignUpParams) => {
    await authService.confirmSignUp(params);
  }, []);

  const resendConfirmationCode = useCallback(async (email: string) => {
    await authService.resendConfirmationCode(email);
  }, []);

  const signIn = useCallback(async (params: SignInParams) => {
    const result = await authService.signIn(params);
    dispatch({ type: 'SIGN_IN', user: result.user });
  }, []);

  const signOut = useCallback(async () => {
    await authService.signOut();
    dispatch({ type: 'SIGN_OUT' });
  }, []);

  const forgotPassword = useCallback(async (params: ForgotPasswordParams) => {
    await authService.forgotPassword(params);
  }, []);

  const resetPassword = useCallback(async (params: ResetPasswordParams) => {
    await authService.resetPassword(params);
  }, []);

  const changePassword = useCallback(async (params: ChangePasswordParams) => {
    await authService.changePassword(params);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      state,
      signUp,
      confirmSignUp,
      resendConfirmationCode,
      signIn,
      signOut,
      forgotPassword,
      resetPassword,
      changePassword,
    }),
    [
      state,
      signUp,
      confirmSignUp,
      resendConfirmationCode,
      signIn,
      signOut,
      forgotPassword,
      resetPassword,
      changePassword,
    ]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
}
