import React from 'react';
import { Modal, View, Text, StyleSheet, type ViewStyle } from 'react-native';
import { Button } from './Button';
import { colors } from '../constants/colors';

export interface ConfirmationDialogProps {
  visible: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmationDialog({
  visible,
  title,
  message,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive = false,
  onConfirm,
  onCancel,
}: ConfirmationDialogProps) {
  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onCancel}
      statusBarTranslucent
    >
      <View style={styles.overlay}>
        <View
          style={styles.dialog}
          accessibilityRole="alert"
          accessibilityLabel={title}
        >
          <Text style={styles.title}>{title}</Text>
          <Text style={styles.message}>{message}</Text>
          <View style={styles.actions}>
            <Button
              title={cancelLabel}
              variant="secondary"
              onPress={onCancel}
              containerStyle={styles.button}
            />
            <Button
              title={confirmLabel}
              variant="primary"
              onPress={onConfirm}
              containerStyle={StyleSheet.flatten([
                styles.button,
                destructive ? styles.destructiveButton : undefined,
              ])}
            />
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  dialog: {
    backgroundColor: colors.primary.white,
    borderRadius: 12,
    padding: 24,
    width: '100%',
    maxWidth: 340,
  },
  title: {
    fontSize: 18,
    fontWeight: '700',
    color: colors.primary.black,
    marginBottom: 8,
  },
  message: {
    fontSize: 14,
    color: colors.primary.greyDark,
    lineHeight: 20,
    marginBottom: 24,
  },
  actions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 12,
  },
  button: {
    flex: 1,
    minHeight: 44,
    paddingVertical: 10,
  },
  destructiveButton: {
    backgroundColor: colors.primary.red,
  },
});
