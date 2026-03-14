import React from 'react';
import { Text, StyleSheet, type ViewStyle } from 'react-native';
import { colors } from '../constants/colors';

export interface FormErrorProps {
  message: string | null | undefined;
  style?: ViewStyle;
}

export function FormError({ message, style }: FormErrorProps) {
  if (!message) {
    return null;
  }

  return (
    <Text
      style={[styles.error, style]}
      accessibilityRole="alert"
      accessibilityLiveRegion="assertive"
    >
      {message}
    </Text>
  );
}

const styles = StyleSheet.create({
  error: {
    color: colors.primary.red,
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 16,
    paddingHorizontal: 16,
  },
});
