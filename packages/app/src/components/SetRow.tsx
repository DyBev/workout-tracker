import { useCallback, useEffect, useRef, useState } from 'react';
import { View, Text, TextInput, Pressable, Platform, StyleSheet } from 'react-native';
import { colors } from '../constants/colors';
import { WorkoutSet } from '../types/workout';
import { ConfirmationDialog } from './ConfirmationDialog';

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

export function SetRow({ set, exerciseId, onUpdateSet, onRemoveSet }: SetRowProps) {
  const [weightInput, setWeightInput] = useState<string>(
    set.weight !== null ? String(set.weight) : '',
  );
  const isWeightFocused = useRef(false);
  const [showConfirmRemoveSet, setShowConfirmRemoveSet] = useState(false);

  useEffect(() => {
    if (!isWeightFocused.current) {
      setWeightInput(set.weight !== null ? String(set.weight) : '');
    }
  }, [set.weight]);

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
      if (text !== '' && isNaN(parseFloat(text)) && text !== '.') {
        return;
      }
      setWeightInput(text);

      if (text === '' || text === '.') {
        onUpdateSet(exerciseId, set.setId, 'weight', text === '' ? null : set.weight);
      } else {
        const parsed = parseFloat(text);
        if (!isNaN(parsed)) {
          onUpdateSet(exerciseId, set.setId, 'weight', parsed);
        }
      }
    },
    [exerciseId, set.setId, set.weight, onUpdateSet],
  );

  const handleWeightBlur = useCallback(() => {
    isWeightFocused.current = false;
    const parsed = parseFloat(weightInput);
    if (weightInput === '' || isNaN(parsed)) {
      setWeightInput('');
      onUpdateSet(exerciseId, set.setId, 'weight', null);
    } else {
      setWeightInput(String(parsed));
      onUpdateSet(exerciseId, set.setId, 'weight', parsed);
    }
  }, [exerciseId, set.setId, weightInput, onUpdateSet]);

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
        value={weightInput}
        onChangeText={handleWeightChange}
        onFocus={() => { isWeightFocused.current = true; }}
        onBlur={handleWeightBlur}
        accessibilityLabel={`Set ${set.order} weight`}
      />
      <Text style={styles.unitLabel}>kg</Text>
      <Pressable
        onPress={() => setShowConfirmRemoveSet(true)}
        accessibilityRole="button"
        accessibilityLabel={`Remove set ${set.order}`}
        style={styles.removeSetButton}
      >
        <Text style={styles.removeSetText}>X</Text>
      </Pressable>

      <ConfirmationDialog
        visible={showConfirmRemoveSet}
        title='Confirm to remove set'
        message={`Remove set ${set.order}`}
        onCancel={() => setShowConfirmRemoveSet(false)}
        onConfirm={handleRemove}
      />

    </View>
  );
}

const styles = StyleSheet.create({
  setRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
	gap: 8,
	flexWrap: 'wrap',
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
});
