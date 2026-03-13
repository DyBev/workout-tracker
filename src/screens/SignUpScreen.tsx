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

type Props = NativeStackScreenProps<AuthStackParamList, 'SignUp'>;

function validateSignUp(values: {
  email: string;
  password: string;
  confirmPassword: string;
}) {
  const errors: Partial<Record<keyof typeof values, string>> = {};

  if (!values.email.trim()) {
    errors.email = 'Email is required';
  } else if (!/\S+@\S+\.\S+/.test(values.email)) {
    errors.email = 'Please enter a valid email address';
  }

  if (!values.password) {
    errors.password = 'Password is required';
  } else if (values.password.length < 8) {
    errors.password = 'Password must be at least 8 characters';
  }

  if (!values.confirmPassword) {
    errors.confirmPassword = 'Please confirm your password';
  } else if (values.password !== values.confirmPassword) {
    errors.confirmPassword = 'Passwords do not match';
  }

  return errors;
}

export function SignUpScreen({ navigation }: Props) {
  const { signUp } = useAuth();
  const passwordRef = useRef<TextInput>(null);
  const confirmPasswordRef = useRef<TextInput>(null);

  const form = useAuthForm({
    initialValues: { email: '', password: '', confirmPassword: '' },
    validate: validateSignUp,
    onSubmit: async (values) => {
      await signUp({ email: values.email.trim(), password: values.password });
      navigation.navigate('ConfirmSignUp', { email: values.email.trim() });
    },
  });

  const handleNavigateToSignIn = useCallback(() => {
    navigation.navigate('SignIn');
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
          accessibilityLabel="Sign up form"
        >
          <Text
            style={styles.title}
            accessibilityRole="header"
          >
            Create Account
          </Text>

          <Text style={styles.subtitle}>
            Sign up to get started.
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
            autoComplete="new-password"
            textContentType="newPassword"
            returnKeyType="next"
            onSubmitEditing={() => confirmPasswordRef.current?.focus()}
            editable={!form.isSubmitting}
          />

          <PasswordInput
            ref={confirmPasswordRef}
            label="Confirm Password"
            required
            value={form.values.confirmPassword}
            onChangeText={(text) => form.setValue('confirmPassword', text)}
            error={form.errors.confirmPassword}
            autoComplete="new-password"
            textContentType="newPassword"
            returnKeyType="done"
            onSubmitEditing={form.handleSubmit}
            editable={!form.isSubmitting}
          />

          <Button
            title="Sign Up"
            onPress={form.handleSubmit}
            loading={form.isSubmitting}
            containerStyle={styles.submitButton}
          />

          <View style={styles.footer}>
            <Text style={styles.footerText}>
              Already have an account?
            </Text>
            <Button
              title="Sign In"
              variant="link"
              onPress={handleNavigateToSignIn}
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
    width: '100%',
    maxWidth: 400,
    padding: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    marginBottom: 8,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    color: '#6b7280',
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
    color: '#6b7280',
    fontSize: 14,
  },
});
