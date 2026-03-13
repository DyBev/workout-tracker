import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
} from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import { FormInput, Button, FormError } from '../components';
import { useAuth } from '../contexts/AuthContext';
import { useAuthForm } from '../hooks/useAuthForm';
import type { AuthStackParamList } from '../types/auth';

type Props = NativeStackScreenProps<AuthStackParamList, 'ConfirmSignUp'>;

function validateConfirmation(values: { code: string }) {
  const errors: Partial<Record<keyof typeof values, string>> = {};

  if (!values.code.trim()) {
    errors.code = 'Verification code is required';
  } else if (!/^\d{6}$/.test(values.code.trim())) {
    errors.code = 'Code must be 6 digits';
  }

  return errors;
}

export function ConfirmSignUpScreen({ navigation, route }: Props) {
  const { email } = route.params;
  const { confirmSignUp, resendConfirmationCode } = useAuth();
  const [isResending, setIsResending] = useState(false);
  const [resendMessage, setResendMessage] = useState<string | null>(null);

  const form = useAuthForm({
    initialValues: { code: '' },
    validate: validateConfirmation,
    onSubmit: async (values) => {
      await confirmSignUp({ email, code: values.code.trim() });
      navigation.navigate('SignIn');
    },
  });

  const handleResendCode = useCallback(async () => {
    setIsResending(true);
    setResendMessage(null);

    try {
      await resendConfirmationCode(email);
      setResendMessage('A new code has been sent to your email.');
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to resend code';
      form.setFormError(message);
    } finally {
      setIsResending(false);
    }
  }, [email, resendConfirmationCode, form]);

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
          accessibilityLabel="Verify your email form"
        >
          <Text
            style={styles.title}
            accessibilityRole="header"
          >
            Verify Email
          </Text>

          <Text style={styles.subtitle}>
            We sent a verification code to{' '}
            <Text style={styles.emailText}>{email}</Text>.
            Enter it below to verify your account.
          </Text>

          <FormError message={form.formError} />

          {resendMessage && (
            <Text
              style={styles.successMessage}
              accessibilityRole="alert"
              accessibilityLiveRegion="polite"
            >
              {resendMessage}
            </Text>
          )}

          <FormInput
            label="Verification Code"
            required
            value={form.values.code}
            onChangeText={(text) => form.setValue('code', text)}
            error={form.errors.code}
            keyboardType="number-pad"
            autoComplete="one-time-code"
            textContentType="oneTimeCode"
            maxLength={6}
            returnKeyType="done"
            onSubmitEditing={form.handleSubmit}
            editable={!form.isSubmitting}
          />

          <Button
            title="Verify"
            onPress={form.handleSubmit}
            loading={form.isSubmitting}
            containerStyle={styles.submitButton}
          />

          <Button
            title="Resend Code"
            variant="link"
            onPress={handleResendCode}
            loading={isResending}
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
  successMessage: {
    color: '#059669',
    fontSize: 14,
    textAlign: 'center',
    marginBottom: 16,
  },
  submitButton: {
    marginTop: 8,
    marginBottom: 16,
  },
});
