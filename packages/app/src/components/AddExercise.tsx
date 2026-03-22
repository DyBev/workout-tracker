import { Button } from './Button';
import {
  TextInput,
  View,
  Text,
  Pressable,
  StyleSheet,
  Platform,
  FlatList,
} from 'react-native';
import { useCallback, useMemo, useState } from 'react';
import { colors } from '../constants/colors';
import { useWorkout } from '../contexts/WorkoutContext';

interface AddExerciseRowProps {
  onAdd: (name: string) => void;
}

export function AddExerciseRow({ onAdd }: AddExerciseRowProps) {
  const { state } = useWorkout();

  const exerciseNames = useMemo(() => {
    const names = new Set<string>();
    for (const w of state.history) {
      for (const ex of w.exercises) {
        names.add(ex.name);
      }
    }
    return Array.from(names).sort((a, b) =>
      a.toLowerCase().localeCompare(b.toLowerCase()),
    );
  }, [state.history]);
  const [name, setName] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);

  const filteredSuggestions = useMemo(() => {
    const trimmed = name.trim().toLowerCase();
    if (!trimmed || exerciseNames.length === 0) return [];
    return exerciseNames.filter((n) => n.toLowerCase().includes(trimmed));
  }, [name, exerciseNames]);

  const handleAdd = useCallback(() => {
    const trimmed = name.trim();
    if (!trimmed) return;
    onAdd(trimmed);
    setName('');
    setShowSuggestions(false);
  }, [name, onAdd]);

  const handleSelect = useCallback(
    (selected: string) => {
      onAdd(selected);
      setName('');
      setShowSuggestions(false);
    },
    [onAdd],
  );

  const handleChangeText = useCallback((text: string) => {
    setName(text);
    setShowSuggestions(true);
  }, []);

  const handleBlur = useCallback(() => {
    // Delay hiding so a tap on a suggestion can register
    setTimeout(() => setShowSuggestions(false), 150);
  }, []);

  const visibleSuggestions = showSuggestions ? filteredSuggestions : [];

  return (
    <View style={styles.wrapper}>
      <View style={styles.addExerciseRow}>
        <View style={styles.inputWrapper}>
          <TextInput
            style={[
              styles.addExerciseInput,
              Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
            ]}
            placeholder="Exercise name"
            placeholderTextColor={colors.primary.grey}
            value={name}
            onChangeText={handleChangeText}
            onSubmitEditing={handleAdd}
            onFocus={() => name.trim() && setShowSuggestions(true)}
            onBlur={handleBlur}
            returnKeyType="done"
            accessibilityLabel="Exercise name"
          />
          {visibleSuggestions.length > 0 && (
            <View style={styles.suggestionsContainer}>
              <FlatList
                data={visibleSuggestions}
                keyExtractor={(item) => item}
                keyboardShouldPersistTaps="handled"
                style={styles.suggestionsList}
                renderItem={({ item }) => (
                  <Pressable
                    style={styles.suggestionItem}
                    onPress={() => handleSelect(item)}
                    accessibilityRole="button"
                    accessibilityLabel={`Select ${item}`}
                  >
                    <Text style={styles.suggestionText}>{item}</Text>
                  </Pressable>
                )}
              />
            </View>
          )}
        </View>
        <Button
          title="Add"
          variant="primary"
          onPress={handleAdd}
          disabled={!name.trim()}
          containerStyle={styles.addExerciseButton}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    marginBottom: 16,
    zIndex: 1,
  },
  addExerciseRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  inputWrapper: {
    flex: 1,
    marginRight: 8,
    position: 'relative',
  },
  addExerciseInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
  },
  addExerciseButton: {
    minHeight: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
  },
  suggestionsContainer: {
    position: 'absolute',
    top: '100%',
    left: 0,
    right: 0,
    backgroundColor: colors.primary.white,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    marginTop: 4,
    maxHeight: 160,
    ...Platform.select({
      ios: {
        shadowColor: colors.primary.black,
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 4,
      },
      android: {
        elevation: 4,
      },
      web: {
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
      },
    }),
  },
  suggestionsList: {
    maxHeight: 160,
  },
  suggestionItem: {
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.primary.greyLight,
  },
  suggestionText: {
    fontSize: 14,
    color: colors.primary.black,
  },
});
