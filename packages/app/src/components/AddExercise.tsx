import { Button } from './Button';
import { TextInput, View, StyleSheet, Platform } from 'react-native';
import { useCallback, useState } from 'react';
import { colors } from '../constants/colors';

interface AddExerciseRowProps {
  onAdd: (name: string) => void;
}

export function AddExerciseRow({ onAdd }: AddExerciseRowProps) {
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

const styles = StyleSheet.create({
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
});
