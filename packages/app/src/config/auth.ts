import { CognitoUserPool } from 'amazon-cognito-identity-js';

const COGNITO_USER_POOL_ID = process.env.EXPO_PUBLIC_COGNITO_USER_POOL_ID ?? '';
const COGNITO_CLIENT_ID = process.env.EXPO_PUBLIC_COGNITO_CLIENT_ID ?? '';

if (!COGNITO_USER_POOL_ID || !COGNITO_CLIENT_ID) {
  console.warn(
    'Cognito configuration is missing. Set EXPO_PUBLIC_COGNITO_USER_POOL_ID and EXPO_PUBLIC_COGNITO_CLIENT_ID in your .env file.'
  );
}

export const cognitoConfig = {
  UserPoolId: COGNITO_USER_POOL_ID,
  ClientId: COGNITO_CLIENT_ID,
} as const;

export const userPool = new CognitoUserPool(cognitoConfig);
