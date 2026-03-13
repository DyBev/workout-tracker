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

type Props = NativeStackScreenProps<AuthStackParamList, 'ResetPassword'>;

function validateResetPassword(values: {
  code: string;
  newPassword: string;
  confirmPassword: string;
}) {
  const errors: Partial<Record<keyof typeof values, string>> = {};

  if (!values.code.trim()) {
    errors.code = 'Reset code is required';
  } else if (!/^\d{6}$/.test(values.code.trim())) {
    errors.code = 'Code must be 6 digits';
  }

  if (!values.newPassword) {
    errors.newPassword = 'New password is required';
  } else if (values.newPassword.length < 8) {
    errors.newPassword = 'Password must be at least 8 characters';
  }

  if (!values.confirmPassword) {
    errors.confirmPassword = 'Please confirm your new password';
  } else if (values.newPassword !== values.confirmPassword) {
    errors.confirmPassword = 'Passwords do not match';
  }

  return errors;
}

export function ResetPasswordScreen({ navigation, route }: Props) {
  const { email } = route.params;
  const { resetPassword } = useAuth();
  const newPasswordRef = useRef<TextInput>(null);
  const confirmPasswordRef = useRef<TextInput>(null);

  const form = useAuthForm({
    initialValues: { code: '', newPassword: '', confirmPassword: '' },
    validate: validateResetPassword,
    onSubmit: async (values) => {
      await resetPassword({
        email,
        code: values.code.trim(),
        newPassword: values.newPassword,
      });
      navigation.navigate('SignIn');
    },
  });

  const handleNavigateBack = useCallback(() => {
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
          accessibilityLabel="Reset password form"
        >
          <Text
            style={styles.title}
            accessibilityRole="header"
          >
            Reset Password
          </Text>

          <Text style={styles.subtitle}>
            Enter the code sent to{' '}
            <Text style={styles.emailText}>{email}</Text>
            {' '}and your new password.
          </Text>

          <FormError message={form.formError} />

          <FormInput
            label="Reset Code"
            required
            value={form.values.code}
            onChangeText={(text) => form.setValue('code', text)}
            error={form.errors.code}
            keyboardType="number-pad"
            autoComplete="one-time-code"
            textContentType="oneTimeCode"
            maxLength={6}
            returnKeyType="next"
            onSubmitEditing={() => newPasswordRef.current?.focus()}
            editable={!form.isSubmitting}
          />

          <PasswordInput
            ref={newPasswordRef}
            label="New Password"
            required
            value={form.values.newPassword}
            onChangeText={(text) => form.setValue('newPassword', text)}
            error={form.errors.newPassword}
            autoComplete="new-password"
            textContentType="newPassword"
            returnKeyType="next"
            onSubmitEditing={() => confirmPasswordRef.current?.focus()}
            editable={!form.isSubmitting}
          />

          <PasswordInput
            ref={confirmPasswordRef}
            label="Confirm New Password"
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
            title="Reset Password"
            onPress={form.handleSubmit}
            loading={form.isSubmitting}
            containerStyle={styles.submitButton}
          />

          <Button
            title="Back to Sign In"
            variant="link"
            onPress={handleNavigateBack}
            disabled={form.isSubmitting}
          />
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
    width: '100%',
    maxWidth: 440,
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
    lineHeight: 22,
  },
  emailText: {
    fontWeight: '600',
    color: '#111827',
  },
  submitButton: {
    marginTop: 8,
    marginBottom: 16,
  },
});
