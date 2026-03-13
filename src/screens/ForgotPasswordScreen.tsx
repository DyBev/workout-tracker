import React, { useCallback } from 'react';
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

type Props = NativeStackScreenProps<AuthStackParamList, 'ForgotPassword'>;

function validateForgotPassword(values: { email: string }) {
  const errors: Partial<Record<keyof typeof values, string>> = {};

  if (!values.email.trim()) {
    errors.email = 'Email is required';
  } else if (!/\S+@\S+\.\S+/.test(values.email)) {
    errors.email = 'Please enter a valid email address';
  }

  return errors;
}

export function ForgotPasswordScreen({ navigation }: Props) {
  const { forgotPassword } = useAuth();

  const form = useAuthForm({
    initialValues: { email: '' },
    validate: validateForgotPassword,
    onSubmit: async (values) => {
      await forgotPassword({ email: values.email.trim() });
      navigation.navigate('ResetPassword', { email: values.email.trim() });
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
          accessibilityLabel="Forgot password form"
        >
          <Text
            style={styles.title}
            accessibilityRole="header"
          >
            Forgot Password
          </Text>

          <Text style={styles.subtitle}>
            Enter your email address and we'll send you a code to reset your
            password.
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
            returnKeyType="done"
            onSubmitEditing={form.handleSubmit}
            editable={!form.isSubmitting}
          />

          <Button
            title="Send Reset Code"
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
  submitButton: {
    marginTop: 8,
    marginBottom: 16,
  },
});
