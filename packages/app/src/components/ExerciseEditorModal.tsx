import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  TextInput,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Platform,
  KeyboardAvoidingView,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../constants/colors';
import { TagInput } from './TagInput';
import type { SavedExercise } from '../types/workout';

interface ExerciseEditorModalProps {
  visible: boolean;
  exercise: SavedExercise;
  onSave: (updated: SavedExercise) => void;
  onCancel: () => void;
}

export function ExerciseEditorModal({
  visible,
  exercise,
  onSave,
  onCancel,
}: ExerciseEditorModalProps) {
  const insets = useSafeAreaInsets();
  const [name, setName] = useState(exercise.name);
  const [note, setNote] = useState(exercise.note);
  const [tags, setTags] = useState<string[]>(exercise.tags);

  // Reset local state when exercise changes
  React.useEffect(() => {
    setName(exercise.name);
    setNote(exercise.note);
    setTags(exercise.tags);
  }, [exercise]);

  const handleSave = useCallback(() => {
    const trimmedName = name.trim();
    if (!trimmedName) return;
    onSave({
      ...exercise,
      name: trimmedName,
      note: note.trim(),
      tags,
    });
  }, [exercise, name, note, tags, onSave]);

  const canSave = name.trim().length > 0;

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="fullScreen"
      onRequestClose={onCancel}
    >
      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <View style={[styles.header, { paddingTop: insets.top + 8 }]}>
          <Pressable
            onPress={onCancel}
            accessibilityRole="button"
            accessibilityLabel="Cancel"
            style={styles.headerButton}
          >
            <Text style={styles.cancelText}>Cancel</Text>
          </Pressable>
          <Text style={styles.headerTitle}>Edit Exercise</Text>
          <Pressable
            onPress={handleSave}
            accessibilityRole="button"
            accessibilityLabel="Save"
            style={styles.headerButton}
            disabled={!canSave}
          >
            <Text
              style={[styles.saveText, !canSave && styles.saveTextDisabled]}
            >
              Save
            </Text>
          </Pressable>
        </View>

        <ScrollView
          style={styles.scrollView}
          contentContainerStyle={[
            styles.scrollContent,
            { paddingBottom: insets.bottom + 24 },
          ]}
          keyboardShouldPersistTaps="handled"
        >
          <Text style={styles.label}>Name</Text>
          <TextInput
            style={[
              styles.nameInput,
              Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
            ]}
            value={name}
            onChangeText={setName}
            placeholder="Exercise name"
            placeholderTextColor={colors.primary.grey}
            autoFocus
            returnKeyType="next"
            accessibilityLabel="Exercise name"
          />

          <Text style={styles.label}>Perpetual Note</Text>
          <TextInput
            style={[
              styles.noteInput,
              Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
            ]}
            value={note}
            onChangeText={setNote}
            placeholder="Add a note that will always show during workouts..."
            placeholderTextColor={colors.primary.grey}
            multiline
            textAlignVertical="top"
            accessibilityLabel="Perpetual note"
          />

          <TagInput tags={tags} onUpdateTags={setTags} />
        </ScrollView>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary.white,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  headerButton: {
    paddingVertical: 8,
    paddingHorizontal: 4,
    minWidth: 60,
  },
  headerTitle: {
    fontSize: 17,
    fontWeight: '700',
  },
  cancelText: {
    fontSize: 16,
    color: colors.primary.blue,
  },
  saveText: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.primary.blue,
    textAlign: 'right',
  },
  saveTextDisabled: {
    opacity: 0.4,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
  },
  nameInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    marginBottom: 24,
  },
  noteInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    minHeight: 120,
    marginBottom: 24,
  },
});
