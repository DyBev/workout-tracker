
import { useCallback } from 'react';
import { WorkoutExercise } from '../types/workout';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { Button } from './Button';
import { SetRow } from './SetRow';
import { colors } from '../constants/colors';

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

export function ExerciseCard({
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

const styles = StyleSheet.create({
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
});
