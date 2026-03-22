import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  TextInput,
  ScrollView,
  Pressable,
  StyleSheet,
  Platform,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useWorkout } from '../contexts/WorkoutContext';
import { useExercise } from '../contexts/ExerciseContext';
import { Button, ConfirmationDialog } from '../components';
import { ErrorDialog } from '../components/ErrorDialog';
import { ExerciseEditorModal } from '../components/ExerciseEditorModal';
import { colors } from '../constants/colors';
import type { ActiveWorkoutScreenProps, SavedExercise } from '../types/workout';
import { TagInput } from '../components/TagInput';
import { AddExerciseRow } from '../components/AddExercise';
import { ExerciseCard } from '../components/ExerciseCard';
import { BodyWeightInput } from '../components/BodyWeightInput';

export function ActiveWorkoutScreen({ navigation }: ActiveWorkoutScreenProps) {
  const {
    state,
    addExercise,
    removeExercise,
    addSet,
    removeSet,
    updateSet,
    updateExerciseName,
    updateNotes,
    updateTags,
    updateBodyWeight,
    completeWorkout,
    discardWorkout,
  } = useWorkout();

  const {
    saveExercise: contextSaveExercise,
    updateExercise: contextUpdateExercise,
    getById,
  } = useExercise();
  const insets = useSafeAreaInsets();
  const [isCompleting, setIsCompleting] = useState(false);
  const [showDiscardDialog, setShowDiscardDialog] = useState(false);
  const [errorDialog, setErrorDialog] = useState<{ title: string; message: string } | null>(null);
  const [showErrorDialog, setShowErrorDialog] = useState<boolean>(false);

  // Exercise editor modal state
  const [editorExercise, setEditorExercise] = useState<SavedExercise | null>(null);
  const [editorWorkoutExerciseId, setEditorWorkoutExerciseId] = useState<string | null>(null);

  const workout = state.activeWorkout;

  const handleComplete = useCallback(async () => {
    if (!workout || workout.exercises.length === 0) {
      setErrorDialog({
        title: 'No exercises',
        message: 'Add at least one exercise before finishing.',
      });
      setShowErrorDialog(true);
      return;
    }
    if (workout.bodyWeight > 1000) {
      setErrorDialog({
        title: 'Invalid body weight',
        message: 'Body weight must be 1000 kg or less.',
      });
      setShowErrorDialog(true);
      return;
    }
    const invalidExercise = workout.exercises.find(
      (ex) => !ex.sets.some((s) => (s.reps ?? 0) > 0 && s.weight !== 0),
    );
    if (invalidExercise) {
      setErrorDialog({
        title: 'Invalid exercise',
        message: `"${invalidExercise.name}" needs at least one set with reps greater than 0 and a weight entered.`,
      });
      setShowErrorDialog(true);
      return;
    }
    setIsCompleting(true);
    try {
      await completeWorkout();
      navigation.replace('WorkoutSummary', {
        workoutId: workout.workoutId,
      });
    } catch {
      setErrorDialog({
        title: 'Error',
        message: 'Failed to save workout. Please try again.',
      });
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

  const handleSaveExercise = useCallback(
    (exerciseId: string, name: string) => {
      const saved = contextSaveExercise(name);
      setEditorExercise(saved);
      setEditorWorkoutExerciseId(exerciseId);
    },
    [contextSaveExercise],
  );

  const handleEditExercise = useCallback(
    (exerciseId: string, savedExercise: SavedExercise) => {
      setEditorExercise(savedExercise);
      setEditorWorkoutExerciseId(exerciseId);
    },
    [],
  );

  const handleEditorSave = useCallback(
    async (updated: SavedExercise) => {
      await contextUpdateExercise(updated);
      if (editorWorkoutExerciseId) {
        updateExerciseName(
          editorWorkoutExerciseId,
          updated.name,
          updated.savedExerciseId,
        );
      }
      setEditorExercise(null);
      setEditorWorkoutExerciseId(null);
    },
    [editorExercise, editorWorkoutExerciseId, contextUpdateExercise, updateExerciseName],
  );

  const handleEditorCancel = useCallback(() => {
    setEditorExercise(null);
    setEditorWorkoutExerciseId(null);
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

        <BodyWeightInput
          value={workout.bodyWeight}
          onChangeValue={updateBodyWeight}
        />

        <TagInput tags={workout.tags} onUpdateTags={updateTags} />

        {workout.exercises.map((exercise) => {
          const saved = getById(exercise.savedExerciseId);
          return (
            <ExerciseCard
              key={exercise.exerciseId}
              exercise={exercise}
              savedExercise={saved}
              onAddSet={addSet}
              onRemoveSet={removeSet}
              onUpdateSet={updateSet}
              onRemoveExercise={removeExercise}
              onSaveExercise={(name) =>
                handleSaveExercise(exercise.exerciseId, name)
              }
              onEditExercise={(savedEx) =>
                handleEditExercise(exercise.exerciseId, savedEx)
              }
            />
          );
        })}

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

      {showErrorDialog && <ErrorDialog
        visible={errorDialog !== null}
        title={errorDialog?.title ?? ''}
        message={errorDialog?.message ?? ''}
        onDismiss={() => {
          setShowErrorDialog(false);
          setErrorDialog(null)
        }}
      />
      }

      {editorExercise && (
        <ExerciseEditorModal
          visible={!!editorExercise}
          exercise={editorExercise}
          onSave={handleEditorSave}
          onCancel={handleEditorCancel}
        />
      )}
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

  emptyText: {
    fontSize: 16,
    color: colors.primary.greyDark,
    textAlign: 'center',
    marginTop: 48,
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
