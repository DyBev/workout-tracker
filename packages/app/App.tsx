import React from 'react';
import { LoadingOverlay } from './src/components/LoadingOverlay';
import { StatusBar } from 'expo-status-bar';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { AuthProvider, useAuth } from './src/contexts/AuthContext';
import { WorkoutProvider } from './src/contexts/WorkoutContext';
import { ExerciseProvider } from './src/contexts/ExerciseContext';
import { SubscriptionProvider } from './src/contexts/SubscriptionContext';
import {
  SignInScreen,
  SignUpScreen,
  ConfirmSignUpScreen,
  ForgotPasswordScreen,
  ResetPasswordScreen,
  HomeScreen,
  ActiveWorkoutScreen,
  WorkoutHistoryScreen,
  WorkoutSummaryScreen,
  SavedExercisesScreen,
} from './src/screens';
import type { AuthStackParamList } from './src/types/auth';
import type { AppStackParamList } from './src/types/workout';

const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const AppStack = createNativeStackNavigator<AppStackParamList>();

function AuthNavigator() {
  return (
    <AuthStack.Navigator
      screenOptions={{ headerShown: false }}
      initialRouteName="SignIn"
    >
      <AuthStack.Screen name="SignIn" component={SignInScreen} />
      <AuthStack.Screen name="SignUp" component={SignUpScreen} />
      <AuthStack.Screen name="ConfirmSignUp" component={ConfirmSignUpScreen} />
      <AuthStack.Screen name="ForgotPassword" component={ForgotPasswordScreen} />
      <AuthStack.Screen name="ResetPassword" component={ResetPasswordScreen} />
    </AuthStack.Navigator>
  );
}

function AppNavigator() {
  return (
    <SubscriptionProvider>
      <ExerciseProvider>
        <WorkoutProvider>
          <AppStack.Navigator screenOptions={{ headerShown: false }}>
            <AppStack.Screen name="Home" component={HomeScreen} />
            <AppStack.Screen name="ActiveWorkout" component={ActiveWorkoutScreen} />
            <AppStack.Screen name="WorkoutHistory" component={WorkoutHistoryScreen} />
            <AppStack.Screen name="WorkoutSummary" component={WorkoutSummaryScreen} />
            <AppStack.Screen name="SavedExercises" component={SavedExercisesScreen} />
          </AppStack.Navigator>
        </WorkoutProvider>
      </ExerciseProvider>
    </SubscriptionProvider>
  );
}

function RootNavigator() {
  const { state } = useAuth();

  if (state.isLoading) {
    return (
      <LoadingOverlay visible={true} message="loading" />
    );
  }

  return state.isAuthenticated ? <AppNavigator /> : <AuthNavigator />;
}

export default function App() {
  return (
    <AuthProvider>
      <NavigationContainer>
        <RootNavigator />
        <StatusBar style="auto" />
      </NavigationContainer>
    </AuthProvider>
  );
}
