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
import { useExercise } from '../contexts/ExerciseContext';

interface ExerciseSuggestion {
  name: string;
  isSaved: boolean;
  savedExerciseId: string | null;
}

interface AddExerciseRowProps {
  onAdd: (name: string, savedExerciseId?: string | null) => void;
}

export function AddExerciseRow({ onAdd }: AddExerciseRowProps) {
  const { state } = useWorkout();
  const { savedExercises } = useExercise();

  const suggestions = useMemo<ExerciseSuggestion[]>(() => {
    const savedNames = new Map<string, string>();
    for (const e of savedExercises) {
      savedNames.set(e.name.toLowerCase(), e.savedExerciseId);
    }

    const historyNames = new Set<string>();
    for (const w of state.history) {
      for (const ex of w.exercises) {
        if (!savedNames.has(ex.name.toLowerCase())) {
          historyNames.add(ex.name);
        }
      }
    }

    const sortAlpha = (a: string, b: string) =>
      a.toLowerCase().localeCompare(b.toLowerCase());

    const savedList: ExerciseSuggestion[] = savedExercises
      .map((e) => e.name)
      .sort(sortAlpha)
      .map((name) => ({
        name,
        isSaved: true,
        savedExerciseId: savedNames.get(name.toLowerCase()) ?? null,
      }));

    const historyList: ExerciseSuggestion[] = Array.from(historyNames)
      .sort(sortAlpha)
      .map((name) => ({
        name,
        isSaved: false,
        savedExerciseId: null,
      }));

    return [...savedList, ...historyList];
  }, [savedExercises, state.history]);

  const [name, setName] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);

  const filteredSuggestions = useMemo(() => {
    const trimmed = name.trim().toLowerCase();
    if (!trimmed || suggestions.length === 0) return [];
    return suggestions.filter((s) => s.name.toLowerCase().includes(trimmed));
  }, [name, suggestions]);

  const handleAdd = useCallback(() => {
    const trimmed = name.trim();
    if (!trimmed) return;
    const match = suggestions.find(
      (s) => s.name.toLowerCase() === trimmed.toLowerCase(),
    );
    onAdd(trimmed, match?.savedExerciseId ?? null);
    setName('');
    setShowSuggestions(false);
  }, [name, onAdd, suggestions]);

  const handleSelect = useCallback(
    (selected: ExerciseSuggestion) => {
      onAdd(selected.name, selected.savedExerciseId);
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
                keyExtractor={(item) => `${item.isSaved ? 's' : 'h'}_${item.name}`}
                keyboardShouldPersistTaps="handled"
                style={styles.suggestionsList}
                renderItem={({ item }) => (
                  <Pressable
                    style={[
                      styles.suggestionItem,
                      item.isSaved && styles.savedSuggestionItem,
                    ]}
                    onPress={() => handleSelect(item)}
                    accessibilityRole="button"
                    accessibilityLabel={`Select ${item.name}${item.isSaved ? ' (saved)' : ''}`}
                  >
                    <View style={styles.suggestionContent}>
                      {item.isSaved && (
                        <View style={styles.bookmarkIcon}>
                          <Text style={styles.bookmarkIconText}>&#9733;</Text>
                        </View>
                      )}
                      <Text
                        style={[
                          styles.suggestionText,
                          item.isSaved && styles.savedSuggestionText,
                        ]}
                      >
                        {item.name}
                      </Text>
                      {item.isSaved && (
                        <View style={styles.savedBadge}>
                          <Text style={styles.savedBadgeText}>Saved</Text>
                        </View>
                      )}
                    </View>
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
    maxHeight: 200,
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
    maxHeight: 200,
  },
  suggestionItem: {
    paddingHorizontal: 12,
    paddingVertical: 10,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.primary.greyLight,
  },
  savedSuggestionItem: {
    backgroundColor: '#f0f5ff',
  },
  suggestionContent: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  bookmarkIcon: {
    marginRight: 6,
  },
  bookmarkIconText: {
    fontSize: 14,
    color: colors.primary.blue,
  },
  suggestionText: {
    fontSize: 14,
    color: colors.primary.black,
    flex: 1,
  },
  savedSuggestionText: {
    fontWeight: '600',
  },
  savedBadge: {
    backgroundColor: colors.primary.blue,
    borderRadius: 8,
    paddingHorizontal: 6,
    paddingVertical: 2,
    marginLeft: 8,
  },
  savedBadgeText: {
    fontSize: 10,
    fontWeight: '700',
    color: colors.primary.white,
  },
});
