import {
  CognitoUser,
  CognitoUserAttribute,
  CognitoUserSession,
  AuthenticationDetails,
} from 'amazon-cognito-identity-js';
import { userPool } from '../config/auth';
import type {
  AuthUser,
  AuthTokens,
  SignUpParams,
  ConfirmSignUpParams,
  SignInParams,
  ForgotPasswordParams,
  ResetPasswordParams,
  ChangePasswordParams,
} from '../types/auth';

function getCognitoUser(email: string): CognitoUser {
  return new CognitoUser({
    Username: email,
    Pool: userPool,
  });
}

function extractUser(session: CognitoUserSession, email: string): AuthUser {
  const idToken = session.getIdToken();
  const payload = idToken.decodePayload();

  return {
    username: payload['cognito:username'] ?? email,
    email: payload['email'] ?? email,
    sub: payload['sub'] ?? '',
  };
}

function extractTokens(session: CognitoUserSession): AuthTokens {
  return {
    idToken: session.getIdToken().getJwtToken(),
    accessToken: session.getAccessToken().getJwtToken(),
    refreshToken: session.getRefreshToken().getToken(),
  };
}

export function signUp({ email, password }: SignUpParams): Promise<void> {
  const attributes = [
    new CognitoUserAttribute({ Name: 'email', Value: email }),
  ];

  return new Promise((resolve, reject) => {
    userPool.signUp(email, password, attributes, [], (error, result) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });
}

export function confirmSignUp({
  email,
  code,
}: ConfirmSignUpParams): Promise<void> {
  const cognitoUser = getCognitoUser(email);

  return new Promise((resolve, reject) => {
    cognitoUser.confirmRegistration(code, true, (error, result) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });
}

export function resendConfirmationCode(email: string): Promise<void> {
  const cognitoUser = getCognitoUser(email);

  return new Promise((resolve, reject) => {
    cognitoUser.resendConfirmationCode((error, result) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });
}

export function signIn({
  email,
  password,
}: SignInParams): Promise<{ user: AuthUser; tokens: AuthTokens }> {
  const cognitoUser = getCognitoUser(email);
  const authDetails = new AuthenticationDetails({
    Username: email,
    Password: password,
  });

  return new Promise((resolve, reject) => {
    cognitoUser.authenticateUser(authDetails, {
      onSuccess: (session: CognitoUserSession) => {
        const user = extractUser(session, email);
        const tokens = extractTokens(session);
        resolve({ user, tokens });
      },
      onFailure: (error: Error) => {
        reject(error);
      },
    });
  });
}

export function signOut(): Promise<void> {
  const cognitoUser = userPool.getCurrentUser();

  return new Promise((resolve) => {
    if (!cognitoUser) {
      resolve();
      return;
    }

    cognitoUser.signOut(() => {
      resolve();
    });
  });
}

export function forgotPassword({ email }: ForgotPasswordParams): Promise<void> {
  const cognitoUser = getCognitoUser(email);

  return new Promise((resolve, reject) => {
    cognitoUser.forgotPassword({
      onSuccess: () => {
        resolve();
      },
      onFailure: (error: Error) => {
        reject(error);
      },
    });
  });
}

export function resetPassword({
  email,
  code,
  newPassword,
}: ResetPasswordParams): Promise<void> {
  const cognitoUser = getCognitoUser(email);

  return new Promise((resolve, reject) => {
    cognitoUser.confirmPassword(code, newPassword, {
      onSuccess: () => {
        resolve();
      },
      onFailure: (error: Error) => {
        reject(error);
      },
    });
  });
}

export function changePassword({
  oldPassword,
  newPassword,
}: ChangePasswordParams): Promise<void> {
  return new Promise((resolve, reject) => {
    const cognitoUser = userPool.getCurrentUser();

    if (!cognitoUser) {
      reject(new Error('No authenticated user'));
      return;
    }

    cognitoUser.getSession(
      (sessionError: Error | null, session: CognitoUserSession | null) => {
        if (sessionError || !session) {
          reject(sessionError ?? new Error('No session'));
          return;
        }

        cognitoUser.changePassword(oldPassword, newPassword, (error, result) => {
          if (error) {
            reject(error);
            return;
          }
          resolve();
        });
      }
    );
  });
}

export function getCurrentSession(): Promise<{
  user: AuthUser;
  tokens: AuthTokens;
} | null> {
  return new Promise((resolve) => {
    const cognitoUser = userPool.getCurrentUser();

    if (!cognitoUser) {
      resolve(null);
      return;
    }

    cognitoUser.getSession(
      (error: Error | null, session: CognitoUserSession | null) => {
        if (error || !session || !session.isValid()) {
          resolve(null);
          return;
        }

        const email =
          session.getIdToken().decodePayload()['email'] ??
          cognitoUser.getUsername();
        const user = extractUser(session, email);
        const tokens = extractTokens(session);
        resolve({ user, tokens });
      }
    );
  });
}
