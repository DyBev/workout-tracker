import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  TextInput,
  ScrollView,
  Pressable,
  Alert,
  StyleSheet,
  Platform,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useWorkout } from '../contexts/WorkoutContext';
import { Button, ConfirmationDialog } from '../components';
import { colors } from '../constants/colors';
import type { ActiveWorkoutScreenProps } from '../types/workout';
import type { WorkoutExercise, WorkoutSet } from '../types/workout';

interface TagInputProps {
  tags: string[];
  onUpdateTags: (tags: string[]) => void;
}

function TagInput({ tags, onUpdateTags }: TagInputProps) {
  const [value, setValue] = useState('');

  const handleAdd = useCallback(() => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (tags.includes(trimmed)) {
      setValue('');
      return;
    }
    onUpdateTags([...tags, trimmed]);
    setValue('');
  }, [value, tags, onUpdateTags]);

  const handleRemove = useCallback(
    (tag: string) => {
      onUpdateTags(tags.filter((t) => t !== tag));
    },
    [tags, onUpdateTags],
  );

  return (
    <View style={styles.tagSection}>
      <Text style={styles.tagLabel}>Tags</Text>
      {tags.length > 0 && (
        <View style={styles.tagList}>
          {tags.map((tag) => (
            <Pressable
              key={tag}
              onPress={() => handleRemove(tag)}
              accessibilityRole="button"
              accessibilityLabel={`Remove tag ${tag}`}
              style={styles.tagPill}
            >
              <Text style={styles.tagPillText}>{tag}</Text>
              <Text style={styles.tagPillRemove}>x</Text>
            </Pressable>
          ))}
        </View>
      )}
      <View style={styles.tagInputRow}>
        <TextInput
          style={[
            styles.tagInput,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          placeholder="e.g. Push, Pull, Legs"
          placeholderTextColor={colors.primary.grey}
          value={value}
          onChangeText={setValue}
          onSubmitEditing={handleAdd}
          returnKeyType="done"
          autoCapitalize="none"
          accessibilityLabel="Add tag"
        />
        <Button
          title="Add"
          variant="secondary"
          onPress={handleAdd}
          disabled={!value.trim()}
          containerStyle={styles.tagAddButton}
        />
      </View>
    </View>
  );
}

interface AddExerciseRowProps {
  onAdd: (name: string) => void;
}

function AddExerciseRow({ onAdd }: AddExerciseRowProps) {
  const [name, setName] = useState('');

  const handleAdd = useCallback(() => {
    const trimmed = name.trim();
    if (!trimmed) return;
    onAdd(trimmed);
    setName('');
  }, [name, onAdd]);

  return (
    <View style={styles.addExerciseRow}>
      <TextInput
        style={[
          styles.addExerciseInput,
          Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
        ]}
        placeholder="Exercise name"
        placeholderTextColor={colors.primary.grey}
        value={name}
        onChangeText={setName}
        onSubmitEditing={handleAdd}
        returnKeyType="done"
        accessibilityLabel="Exercise name"
      />
      <Button
        title="Add"
        variant="primary"
        onPress={handleAdd}
        disabled={!name.trim()}
        containerStyle={styles.addExerciseButton}
      />
    </View>
  );
}

interface SetRowProps {
  set: WorkoutSet;
  exerciseId: string;
  onUpdateSet: (
    exerciseId: string,
    setId: string,
    field: 'reps' | 'weight',
    value: number | null,
  ) => void;
  onRemoveSet: (exerciseId: string, setId: string) => void;
}

function SetRow({ set, exerciseId, onUpdateSet, onRemoveSet }: SetRowProps) {
  const handleRepsChange = useCallback(
    (text: string) => {
      const parsed = text === '' ? null : parseInt(text, 10);
      const value = parsed !== null && isNaN(parsed) ? set.reps : parsed;
      onUpdateSet(exerciseId, set.setId, 'reps', value);
    },
    [exerciseId, set.setId, set.reps, onUpdateSet],
  );

  const handleWeightChange = useCallback(
    (text: string) => {
      const parsed = text === '' ? null : parseFloat(text);
      const value = parsed !== null && isNaN(parsed) ? set.weight : parsed;
      onUpdateSet(exerciseId, set.setId, 'weight', value);
    },
    [exerciseId, set.setId, set.weight, onUpdateSet],
  );

  const handleRemove = useCallback(() => {
    onRemoveSet(exerciseId, set.setId);
  }, [exerciseId, set.setId, onRemoveSet]);

  return (
    <View style={styles.setRow}>
      <Text style={styles.setLabel} accessibilityLabel={`Set ${set.order}`}>
        {set.order}
      </Text>
      <TextInput
        style={[
          styles.setInput,
          Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
        ]}
        placeholder="Reps"
        placeholderTextColor={colors.primary.grey}
        keyboardType="number-pad"
        value={set.reps !== null ? String(set.reps) : ''}
        onChangeText={handleRepsChange}
        accessibilityLabel={`Set ${set.order} reps`}
      />
      <TextInput
        style={[
          styles.setInput,
          Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
        ]}
        placeholder="Weight"
        placeholderTextColor={colors.primary.grey}
        keyboardType="decimal-pad"
        value={set.weight !== null ? String(set.weight) : ''}
        onChangeText={handleWeightChange}
        accessibilityLabel={`Set ${set.order} weight`}
      />
      <Text style={styles.unitLabel}>kg</Text>
      <Pressable
        onPress={handleRemove}
        accessibilityRole="button"
        accessibilityLabel={`Remove set ${set.order}`}
        style={styles.removeSetButton}
      >
        <Text style={styles.removeSetText}>X</Text>
      </Pressable>
    </View>
  );
}

interface ExerciseCardProps {
  exercise: WorkoutExercise;
  onAddSet: (exerciseId: string) => void;
  onRemoveSet: (exerciseId: string, setId: string) => void;
  onUpdateSet: (
    exerciseId: string,
    setId: string,
    field: 'reps' | 'weight',
    value: number | null,
  ) => void;
  onRemoveExercise: (exerciseId: string) => void;
}

function ExerciseCard({
  exercise,
  onAddSet,
  onRemoveSet,
  onUpdateSet,
  onRemoveExercise,
}: ExerciseCardProps) {
  const handleAddSet = useCallback(() => {
    onAddSet(exercise.exerciseId);
  }, [exercise.exerciseId, onAddSet]);

  const handleRemoveExercise = useCallback(() => {
    onRemoveExercise(exercise.exerciseId);
  }, [exercise.exerciseId, onRemoveExercise]);

  return (
    <View style={styles.exerciseCard}>
      <View style={styles.exerciseHeader}>
        <Text
          style={styles.exerciseName}
          accessibilityRole="header"
        >
          {exercise.name}
        </Text>
        <Pressable
          onPress={handleRemoveExercise}
          accessibilityRole="button"
          accessibilityLabel={`Remove ${exercise.name}`}
          style={styles.removeExerciseButton}
        >
          <Text style={styles.removeExerciseText}>Remove</Text>
        </Pressable>
      </View>

      {exercise.sets.length > 0 && (
        <View style={styles.setHeaderRow}>
          <Text style={[styles.setHeaderText, styles.setLabel]}>Set</Text>
          <Text style={[styles.setHeaderText, styles.setInput]}>Reps</Text>
          <Text style={[styles.setHeaderText, styles.setInput]}>Weight</Text>
          <Text style={[styles.setHeaderText, styles.unitLabel]} />
          <View style={styles.removeSetButton} />
        </View>
      )}

      {exercise.sets.map((set) => (
        <SetRow
          key={set.setId}
          set={set}
          exerciseId={exercise.exerciseId}
          onUpdateSet={onUpdateSet}
          onRemoveSet={onRemoveSet}
        />
      ))}

      <Button
        title="+ Add Set"
        variant="link"
        onPress={handleAddSet}
        containerStyle={styles.addSetButton}
      />
    </View>
  );
}

export function ActiveWorkoutScreen({ navigation }: ActiveWorkoutScreenProps) {
  const {
    state,
    addExercise,
    removeExercise,
    addSet,
    removeSet,
    updateSet,
    updateNotes,
    updateTags,
    completeWorkout,
    discardWorkout,
  } = useWorkout();
  const insets = useSafeAreaInsets();
  const [isCompleting, setIsCompleting] = useState(false);
  const [showDiscardDialog, setShowDiscardDialog] = useState(false);

  const workout = state.activeWorkout;

  const handleComplete = useCallback(async () => {
    if (!workout || workout.exercises.length === 0) {
      Alert.alert(
        'No exercises',
        'Add at least one exercise before finishing.',
      );
      return;
    }
    setIsCompleting(true);
    try {
      await completeWorkout();
      navigation.replace('WorkoutSummary', {
        workoutId: workout.workoutId,
      });
    } catch {
      Alert.alert('Error', 'Failed to save workout. Please try again.');
      setIsCompleting(false);
    }
  }, [workout, completeWorkout, navigation]);

  const handleDiscard = useCallback(() => {
    setShowDiscardDialog(true);
  }, []);

  const handleConfirmDiscard = useCallback(async () => {
    setShowDiscardDialog(false);
    await discardWorkout();
    navigation.navigate('Home');
  }, [discardWorkout, navigation]);

  const handleCancelDiscard = useCallback(() => {
    setShowDiscardDialog(false);
  }, []);

  if (!workout) {
    return (
      <View style={[styles.container, { paddingTop: insets.top }]}>
        <Text style={styles.emptyText}>No active workout.</Text>
      </View>
    );
  }

  const elapsed = formatElapsed(workout.startedAt);

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <View style={styles.header}>
        <View>
          <Text style={styles.title} accessibilityRole="header">
            Workout
          </Text>
          <Text style={styles.elapsed}>Started {elapsed}</Text>
        </View>
        <Pressable
          onPress={handleDiscard}
          accessibilityRole="button"
          accessibilityLabel="Discard workout"
          style={styles.discardButton}
        >
          <Text style={styles.discardText}>Discard</Text>
        </Pressable>
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={[
          styles.scrollContent,
          { paddingBottom: insets.bottom + 80 },
        ]}
        keyboardShouldPersistTaps="handled"
      >
        <TextInput
          style={[
            styles.notesInput,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          placeholder="Workout notes (optional)"
          placeholderTextColor={colors.primary.grey}
          value={workout.notes}
          onChangeText={updateNotes}
          multiline
          accessibilityLabel="Workout notes"
        />

        <TagInput tags={workout.tags} onUpdateTags={updateTags} />

        {workout.exercises.map((exercise) => (
          <ExerciseCard
            key={exercise.exerciseId}
            exercise={exercise}
            onAddSet={addSet}
            onRemoveSet={removeSet}
            onUpdateSet={updateSet}
            onRemoveExercise={removeExercise}
          />
        ))}

        <AddExerciseRow onAdd={addExercise} />
      </ScrollView>

      <View
        style={[styles.finishBar, { paddingBottom: Math.max(insets.bottom, 16) }]}
      >
        <Button
          title="Finish Workout"
          variant="primary"
          onPress={handleComplete}
          loading={isCompleting}
          containerStyle={styles.finishButton}
        />
      </View>

      <ConfirmationDialog
        visible={showDiscardDialog}
        title="Discard workout?"
        message="All progress will be lost."
        confirmLabel="Discard"
        cancelLabel="Cancel"
        destructive
        onConfirm={handleConfirmDiscard}
        onCancel={handleCancelDiscard}
      />
    </View>
  );
}

function formatElapsed(isoString: string): string {
  const start = new Date(isoString).getTime();
  const now = Date.now();
  const diffMs = now - start;

  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  return `${hours}h ${remainingMinutes}m ago`;
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
    paddingHorizontal: 24,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
  },
  elapsed: {
    fontSize: 13,
    color: colors.primary.greyDark,
    marginTop: 2,
  },
  discardButton: {
    paddingVertical: 8,
    paddingHorizontal: 12,
  },
  discardText: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.red,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
  },
  notesInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    minHeight: 48,
    marginBottom: 16,
  },
  tagSection: {
    marginBottom: 24,
  },
  tagLabel: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
  },
  tagList: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 8,
  },
  tagPill: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.primary.blue,
    borderRadius: 16,
    paddingVertical: 6,
    paddingHorizontal: 12,
    marginRight: 8,
    marginBottom: 8,
  },
  tagPillText: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.primary.white,
  },
  tagPillRemove: {
    fontSize: 13,
    fontWeight: '700',
    color: colors.primary.white,
    marginLeft: 6,
    opacity: 0.8,
  },
  tagInputRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  tagInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    marginRight: 8,
  },
  tagAddButton: {
    minHeight: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
  },

  emptyText: {
    fontSize: 16,
    color: colors.primary.greyDark,
    textAlign: 'center',
    marginTop: 48,
  },
  exerciseCard: {
    backgroundColor: colors.primary.white,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    padding: 16,
    marginBottom: 16,
  },
  exerciseHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  exerciseName: {
    fontSize: 16,
    fontWeight: '700',
    flex: 1,
  },
  removeExerciseButton: {
    paddingVertical: 4,
    paddingHorizontal: 8,
  },
  removeExerciseText: {
    fontSize: 13,
    color: colors.primary.red,
    fontWeight: '600',
  },
  setHeaderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  setHeaderText: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.primary.greyDark,
  },
  setRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  setLabel: {
    width: 32,
    textAlign: 'center',
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.greyDark,
  },
  setInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 6,
    paddingHorizontal: 10,
    paddingVertical: 8,
    fontSize: 14,
    marginHorizontal: 4,
    textAlign: 'center',
  },
  unitLabel: {
    width: 28,
    fontSize: 13,
    color: colors.primary.greyDark,
    textAlign: 'center',
  },
  removeSetButton: {
    width: 32,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 4,
  },
  removeSetText: {
    fontSize: 14,
    fontWeight: '700',
    color: colors.primary.red,
  },
  addSetButton: {
    alignSelf: 'flex-start',
    marginTop: 4,
  },
  addExerciseRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  addExerciseInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    marginRight: 8,
  },
  addExerciseButton: {
    minHeight: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
  },
  finishBar: {
    paddingHorizontal: 24,
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: colors.primary.greyLight,
    backgroundColor: colors.primary.white,
  },
  finishButton: {
    width: '100%',
  },
});
