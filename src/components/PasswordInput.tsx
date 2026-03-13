import React, { forwardRef, useState } from 'react';
import {
  TextInput,
  Text,
  View,
  Pressable,
  StyleSheet,
  type ViewStyle,
  type TextInputProps,
  Platform,
} from 'react-native';

export interface PasswordInputProps extends Omit<TextInputProps, 'secureTextEntry'> {
  label: string;
  error?: string;
  containerStyle?: ViewStyle;
  required?: boolean;
}

export const PasswordInput = forwardRef<TextInput, PasswordInputProps>(
  function PasswordInput(
    { label, error, containerStyle, required = false, ...textInputProps },
    ref
  ) {
    const [isVisible, setIsVisible] = useState(false);

    const inputId = `input-${label.toLowerCase().replace(/\s+/g, '-')}`;
    const errorId = `${inputId}-error`;
    const hasError = !!error;

    const toggleVisibility = () => {
      setIsVisible((prev) => !prev);
    };

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
        <View style={[styles.inputWrapper, hasError && styles.inputError]}>
          <TextInput
            ref={ref}
            style={[
              styles.input,
              Platform.OS === 'web' && { outlineStyle: 'none' },
            ]}
            secureTextEntry={!isVisible}
            accessibilityLabelledBy={inputId}
            accessibilityLabel={label}
            accessibilityState={{ disabled: textInputProps.editable === false }}
            aria-required={required}
            aria-invalid={hasError}
            aria-errormessage={hasError ? errorId : undefined}
            autoCorrect={false}
            autoCapitalize="none"
            textContentType="password"
            {...textInputProps}
          />
          <Pressable
            onPress={toggleVisibility}
            style={styles.toggleButton}
            accessibilityRole="button"
            accessibilityLabel={isVisible ? 'Hide password' : 'Show password'}
            accessibilityState={{ checked: isVisible }}
            hitSlop={8}
          >
            <Text style={styles.toggleText}>
              {isVisible ? 'Hide' : 'Show'}
            </Text>
          </Pressable>
        </View>
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
  inputWrapper: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#d1d5db',
    borderRadius: 8,
  },
  inputError: {
    borderColor: '#dc2626',
  },
  input: {
    flex: 1,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
  },
  toggleButton: {
    paddingHorizontal: 12,
    paddingVertical: 10,
  },
  toggleText: {
    color: '#6b7280',
    fontSize: 14,
    fontWeight: '500',
  },
  errorText: {
    color: '#dc2626',
    fontSize: 12,
    marginTop: 4,
  },
});
