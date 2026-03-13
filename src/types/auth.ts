import type { CognitoUserSession } from 'amazon-cognito-identity-js';

export interface AuthUser {
  username: string;
  email: string;
  sub: string;
}

export interface AuthState {
  user: AuthUser | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

export interface AuthTokens {
  idToken: string;
  accessToken: string;
  refreshToken: string;
}

export interface SignUpParams {
  email: string;
  password: string;
}

export interface ConfirmSignUpParams {
  email: string;
  code: string;
}

export interface SignInParams {
  email: string;
  password: string;
}

export interface ForgotPasswordParams {
  email: string;
}

export interface ResetPasswordParams {
  email: string;
  code: string;
  newPassword: string;
}

export interface ChangePasswordParams {
  oldPassword: string;
  newPassword: string;
}

export interface AuthContextValue {
  state: AuthState;
  signUp: (params: SignUpParams) => Promise<void>;
  confirmSignUp: (params: ConfirmSignUpParams) => Promise<void>;
  resendConfirmationCode: (email: string) => Promise<void>;
  signIn: (params: SignInParams) => Promise<void>;
  signOut: () => Promise<void>;
  forgotPassword: (params: ForgotPasswordParams) => Promise<void>;
  resetPassword: (params: ResetPasswordParams) => Promise<void>;
  changePassword: (params: ChangePasswordParams) => Promise<void>;
}

export type AuthScreen =
  | 'SignIn'
  | 'SignUp'
  | 'ConfirmSignUp'
  | 'ForgotPassword'
  | 'ResetPassword';

export type AuthStackParamList = {
  SignIn: undefined;
  SignUp: undefined;
  ConfirmSignUp: { email: string };
  ForgotPassword: undefined;
  ResetPassword: { email: string };
};

export type AppStackParamList = {
  Home: undefined;
};

export type FormField = {
  value: string;
  error: string;
};
