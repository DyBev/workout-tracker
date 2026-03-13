import React, { forwardRef, useState } from 'react';
import {
  TextInput,
  Text,
  View,
  StyleSheet,
  type TextInputProps,
  type ViewStyle,
  Platform,
} from 'react-native';

export interface FormInputProps extends TextInputProps {
  label: string;
  error?: string;
  containerStyle?: ViewStyle;
  required?: boolean;
}

export const FormInput = forwardRef<TextInput, FormInputProps>(
  function FormInput(
    { label, error, containerStyle, required = false, ...textInputProps },
    ref
  ) {
    const inputId = `input-${label.toLowerCase().replace(/\s+/g, '-')}`;
    const errorId = `${inputId}-error`;
    const hasError = !!error;

    return (
      <View
        style={[styles.container, containerStyle]}
        accessible={false}
      >
        <Text
          style={styles.label}
          nativeID={inputId}
        >
          {label}
          {required && (
            <Text
              style={styles.required}
              accessibilityLabel="required"
            >
              {' *'}
            </Text>
          )}
        </Text>
        <TextInput
          ref={ref}
          style={[
            styles.input,
            hasError && styles.inputError,
            Platform.OS === 'web' && { outlineStyle: 'none' },
          ]}
          accessibilityLabelledBy={inputId}
          accessibilityLabel={label}
          accessibilityState={{ disabled: textInputProps.editable === false }}
          aria-required={required}
          aria-invalid={hasError}
          aria-errormessage={hasError ? errorId : undefined}
          autoCorrect={false}
          {...textInputProps}
        />
        {hasError && (
          <Text
            style={styles.errorText}
            nativeID={errorId}
            accessibilityRole="alert"
            accessibilityLiveRegion="polite"
          >
            {error}
          </Text>
        )}
      </View>
    );
  }
);

const styles = StyleSheet.create({
  container: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 6,
  },
  required: {
    color: '#dc2626',
  },
  input: {
    borderWidth: 1,
    borderColor: '#d1d5db',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
  },
  inputError: {
    borderColor: '#dc2626',
  },
  errorText: {
    color: '#dc2626',
    fontSize: 12,
    marginTop: 4,
  },
});
