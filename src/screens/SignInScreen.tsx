import React, { useCallback, useRef } from 'react';
import {
  View,
  Text,
  TextInput,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { FormInput, PasswordInput, Button, FormError } from '../components';
import { useAuth } from '../contexts/AuthContext';
import { useAuthForm } from '../hooks/useAuthForm';
import type { AuthStackParamList } from '../types/auth';
import { colors } from '../constants/colors';

type Props = NativeStackScreenProps<AuthStackParamList, 'SignIn'>;

function validateSignIn(values: { email: string; password: string }) {
  const errors: Partial<Record<keyof typeof values, string>> = {};

  if (!values.email.trim()) {
    errors.email = 'Email is required';
  } else if (!/\S+@\S+\.\S+/.test(values.email)) {
    errors.email = 'Please enter a valid email address';
  }

  if (!values.password) {
    errors.password = 'Password is required';
  }

  return errors;
}

export function SignInScreen({ navigation }: Props) {
  const { signIn } = useAuth();
  const passwordRef = useRef<TextInput>(null);

  const form = useAuthForm({
    initialValues: { email: '', password: '' },
    validate: validateSignIn,
    onSubmit: async (values) => {
      await signIn({ email: values.email.trim(), password: values.password });
    },
  });

  const handleNavigateToSignUp = useCallback(() => {
    navigation.navigate('SignUp');
  }, [navigation]);

  const handleNavigateToForgotPassword = useCallback(() => {
    navigation.navigate('ForgotPassword');
  }, [navigation]);

  return (
    <KeyboardAvoidingView
      style={styles.keyboardView}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView
        contentContainerStyle={styles.scrollContent}
        keyboardShouldPersistTaps="handled"
      >
        <View
          style={styles.container}
          accessible={true}
          accessibilityLabel="Sign in form"
        >
          <Text
            style={styles.title}
            accessibilityRole="header"
          >
            Sign In
          </Text>

          <Text style={styles.subtitle}>
            Welcome back. Sign in to continue.
          </Text>

          <FormError message={form.formError} />

          <FormInput
            label="Email"
            required
            value={form.values.email}
            onChangeText={(text) => form.setValue('email', text)}
            error={form.errors.email}
            keyboardType="email-address"
            autoCapitalize="none"
            autoComplete="email"
            textContentType="emailAddress"
            returnKeyType="next"
            onSubmitEditing={() => passwordRef.current?.focus()}
            editable={!form.isSubmitting}
          />

          <PasswordInput
            ref={passwordRef}
            label="Password"
            required
            value={form.values.password}
            onChangeText={(text) => form.setValue('password', text)}
            error={form.errors.password}
            autoComplete="password"
            returnKeyType="done"
            onSubmitEditing={form.handleSubmit}
            editable={!form.isSubmitting}
          />

          <Button
            title="Sign In"
            onPress={form.handleSubmit}
            loading={form.isSubmitting}
            containerStyle={styles.submitButton}
          />

          <Button
            title="Forgot password?"
            variant="link"
            onPress={handleNavigateToForgotPassword}
            disabled={form.isSubmitting}
          />

          <View style={styles.footer}>
            <Text style={styles.footerText}>
              Don't have an account?
            </Text>
            <Button
              title="Sign Up"
              variant="link"
              onPress={handleNavigateToSignUp}
              disabled={form.isSubmitting}
            />
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  keyboardView: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  container: {
    padding: 24,
    maxWidth: 440,
    width: '100%',
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    marginBottom: 8,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    color: colors.primary.greyDarkest,
    marginBottom: 32,
    textAlign: 'center',
  },
  submitButton: {
    marginTop: 8,
    marginBottom: 16,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 16,
  },
  footerText: {
    color: colors.primary.greyDarkest,
    fontSize: 14,
  },
});
