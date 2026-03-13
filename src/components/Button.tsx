import React from 'react';
import {
  Pressable,
  Text,
  ActivityIndicator,
  StyleSheet,
  type ViewStyle,
  type PressableProps,
} from 'react-native';

export type ButtonVariant = 'primary' | 'secondary' | 'link';

export interface ButtonProps extends Omit<PressableProps, 'children'> {
  title: string;
  variant?: ButtonVariant;
  loading?: boolean;
  containerStyle?: ViewStyle;
}

export function Button({
  title,
  variant = 'primary',
  loading = false,
  disabled,
  containerStyle,
  ...pressableProps
}: ButtonProps) {
  const isDisabled = disabled || loading;

  const buttonStyles = [
    styles.base,
    variant === 'primary' && styles.primary,
    variant === 'secondary' && styles.secondary,
    variant === 'link' && styles.link,
    isDisabled && styles.disabled,
    containerStyle,
  ];

  const textStyles = [
    styles.text,
    variant === 'primary' && styles.primaryText,
    variant === 'secondary' && styles.secondaryText,
    variant === 'link' && styles.linkText,
    isDisabled && styles.disabledText,
  ];

  return (
    <Pressable
      style={buttonStyles}
      disabled={isDisabled}
      accessibilityRole="button"
      accessibilityLabel={title}
      accessibilityState={{
        disabled: isDisabled,
        busy: loading,
      }}
      {...pressableProps}
    >
      {loading ? (
        <ActivityIndicator
          size="small"
          color={variant === 'primary' ? '#ffffff' : '#6b7280'}
          accessibilityLabel="Loading"
        />
      ) : (
        <Text style={textStyles}>{title}</Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },
  primary: {
    backgroundColor: '#2563eb',
  },
  secondary: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#2563eb',
  },
  link: {
    backgroundColor: 'transparent',
    paddingVertical: 8,
    paddingHorizontal: 8,
    minHeight: 0,
  },
  disabled: {
    opacity: 0.5,
  },
  text: {
    fontSize: 16,
    fontWeight: '600',
  },
  primaryText: {
    color: '#ffffff',
  },
  secondaryText: {
    color: '#2563eb',
  },
  linkText: {
    color: '#2563eb',
    fontWeight: '400',
  },
  disabledText: {
    color: '#9ca3af',
  },
});
