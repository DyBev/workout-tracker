import { useCallback, useState, useEffect } from 'react';
import { WorkoutExercise, SavedExercise } from '../types/workout';
import { View, Text, Pressable, Modal, StyleSheet, Platform, TextInput } from 'react-native';
import { Button } from './Button';
import { SetRow } from './SetRow';
import { colors } from '../constants/colors';
import { DocumentAttachment, Star, StarFilled } from '@carbon/icons-react';

interface ExerciseCardProps {
  exercise: WorkoutExercise;
  savedExercise: SavedExercise | undefined;
  onAddSet: (exerciseId: string) => void;
  onRemoveSet: (exerciseId: string, setId: string) => void;
  onUpdateSet: (
    exerciseId: string,
    setId: string,
    field: 'reps' | 'weight',
    value: number | null,
  ) => void;
  onRemoveExercise: (exerciseId: string) => void;
  onSaveExercise: (name: string) => void;
  onEditExercise: (savedExercise: SavedExercise) => void;
  onUpdateNote: (exerciseId: string, note: string | null) => void;
}

export function ExerciseCard({
  exercise,
  savedExercise,
  onAddSet,
  onRemoveSet,
  onUpdateSet,
  onRemoveExercise,
  onSaveExercise,
  onEditExercise,
  onUpdateNote,
}: ExerciseCardProps) {
  const [showNotePopover, setShowNotePopover] = useState(false);
  const [showNoteInput, setShowNoteInput] = useState(false);
  const [noteText, setNoteText] = useState(exercise.note ?? '');

  useEffect(() => {
    setNoteText(exercise.note ?? '');
  }, [exercise.note]);

  const handleAddSet = useCallback(() => {
    onAddSet(exercise.exerciseId);
  }, [exercise.exerciseId, onAddSet]);

  const handleRemoveExercise = useCallback(() => {
    onRemoveExercise(exercise.exerciseId);
  }, [exercise.exerciseId, onRemoveExercise]);

  const handleBookmarkPress = useCallback(() => {
    if (savedExercise) {
      onEditExercise(savedExercise);
    } else {
      onSaveExercise(exercise.name);
    }
  }, [savedExercise, onEditExercise, onSaveExercise, exercise.name]);

  const handleNotePress = useCallback(() => {
    setShowNotePopover(true);
  }, []);

  const isSaved = !!savedExercise;
  const hasNote = !!(savedExercise?.note);

  return (
    <View style={styles.exerciseCard}>
      <View style={styles.exerciseHeader}>
        <Pressable
          onPress={handleBookmarkPress}
          accessibilityRole="button"
          accessibilityLabel={isSaved ? `Edit saved exercise ${exercise.name}` : `Save exercise ${exercise.name}`}
          style={styles.bookmarkButton}
        >
          { isSaved ? <StarFilled size={16} />  : <Star size={16} /> }
        </Pressable>
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
      {hasNote && (
        <Pressable
          onPress={handleNotePress}
          accessibilityRole="button"
          accessibilityLabel={`View note for ${exercise.name}`}
          style={styles.noteButton}
        >
          <DocumentAttachment size={16} /> <Text>Saved note</Text>
        </Pressable>
      )}

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

      <View style={styles.buttonRow}>
        <Button
          title="+ Add Set"
          variant="link"
          onPress={handleAddSet}
          containerStyle={styles.addSetButton}
        />
        {!showNoteInput && (
          <Button
            title="+ Add Note"
            variant="link"
            onPress={() => setShowNoteInput(true)}
            containerStyle={styles.addSetButton}
          />
        )}
      </View>

      {showNoteInput && (
        <TextInput
          style={[
            styles.noteInput,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          value={noteText}
          onChangeText={(t) => {
            setNoteText(t);
            // persist immediately so parent save includes the latest text
            onUpdateNote(exercise.exerciseId, t === '' ? null : t);
          }}
          placeholder="Add a note..."
          placeholderTextColor={colors.primary.grey}
          autoFocus
          onBlur={() => {
            const trimmed = noteText ? noteText.trim() : '';
            if (!trimmed) {
              setNoteText('');
              setShowNoteInput(false);
              onUpdateNote(exercise.exerciseId, null);
              return;
            }
            onUpdateNote(exercise.exerciseId, trimmed);
            setNoteText(trimmed);
            setShowNoteInput(false);
          }}
          accessibilityLabel={`Add note for ${exercise.name}`}
        />
      )}

      <Modal
        visible={showNotePopover}
        transparent
        animationType="fade"
        onRequestClose={() => setShowNotePopover(false)}
      >
        <Pressable
          style={styles.popoverOverlay}
          onPress={() => setShowNotePopover(false)}
        >
          <View style={styles.popoverContainer}>
            <View style={styles.popoverHeader}>
              <Text style={styles.popoverTitle}>Note - {exercise.name}</Text>
              <Pressable
                onPress={() => setShowNotePopover(false)}
                accessibilityRole="button"
                accessibilityLabel="Close note"
                style={styles.popoverClose}
              >
                <Text style={styles.popoverCloseText}>Close</Text>
              </Pressable>
            </View>
            <Text style={styles.popoverText}>{savedExercise.note}</Text>
          </View>
        </Pressable>
      </Modal>
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
    alignItems: 'center',
    marginBottom: 12,
  },
  bookmarkButton: {
    paddingRight: 8,
    paddingVertical: 2,
  },
  bookmarkIcon: {
    fontSize: 18,
    color: colors.primary.grey,
  },
  bookmarkIconSaved: {
    color: colors.primary.blue,
  },
  exerciseName: {
    fontSize: 16,
    fontWeight: '700',
    flex: 1,
  },
  noteButton: {
    paddingHorizontal: '1rem',
    paddingVertical: '1rem',
    display: 'flex',
    flexDirection: 'row',
    gap: 8,
    alignItems: 'center',
    color: colors.primary.blue,
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
  buttonRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 4,
  },
  noteInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 6,
    paddingHorizontal: 10,
    paddingVertical: 8,
    fontSize: 14,
    marginTop: 8,
    
  },
  
  popoverOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.4)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  popoverContainer: {
    backgroundColor: colors.primary.white,
    borderRadius: 12,
    padding: 20,
    width: '100%',
    maxWidth: 400,
    maxHeight: '60%',
    ...Platform.select({
      ios: {
        shadowColor: colors.primary.black,
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.15,
        shadowRadius: 12,
      },
      android: {
        elevation: 8,
      },
      web: {
        boxShadow: '0 4px 24px rgba(0,0,0,0.15)',
      },
    }),
  },
  popoverHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  popoverTitle: {
    fontSize: 16,
    fontWeight: '700',
    flex: 1,
  },
  popoverClose: {
    paddingVertical: 4,
    paddingHorizontal: 8,
  },
  popoverCloseText: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.blue,
  },
  popoverText: {
    fontSize: 14,
    color: colors.primary.greyDarkest,
    lineHeight: 20,
  },
});
