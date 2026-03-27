import React, { useCallback, useMemo, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  Pressable,
  StyleSheet,
  Alert,
  Modal,
  Platform,
} from 'react-native';
import { DocumentAttachment } from '@carbon/icons-react';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useWorkout } from '../contexts/WorkoutContext';
import { colors } from '../constants/colors';
import type { WorkoutSummaryScreenProps } from '../types/workout';

export function WorkoutSummaryScreen({
  navigation,
  route,
}: WorkoutSummaryScreenProps) {
  const { state } = useWorkout();
  const insets = useSafeAreaInsets();
  const { workoutId } = route.params;
  const [showNote, setShowNote] = useState<boolean>(false);
  const [note, setNote] = useState("");
  const [noteExercise, setNoteExercise] = useState("");

  const workout = useMemo(
    () => state.history.find((w) => w.workoutId === workoutId) ?? null,
    [state.history, workoutId],
  );

  const handleDone = useCallback(() => {
    navigation.navigate('WorkoutHistory');
  }, [navigation]);

  if (!workout) {
    return (
      <View style={[styles.container, { paddingTop: insets.top }]}>
        <Text style={styles.emptyText}>Workout not found.</Text>
        <Pressable
          onPress={handleDone}
          accessibilityRole="button"
          accessibilityLabel="Go to home"
          style={styles.doneButton}
        >
          <Text style={styles.doneText}>Go Home</Text>
        </Pressable>
      </View>
    );
  }

  const totalSets = workout.exercises.reduce(
    (sum, ex) => sum + ex.sets.length,
    0,
  );
  const totalReps = workout.exercises.reduce(
    (sum, ex) =>
      sum + ex.sets.reduce((setSum, s) => setSum + (s.reps ?? 0), 0),
    0,
  );
  const totalVolume = workout.exercises.reduce(
    (sum, ex) =>
      sum +
      ex.sets.reduce(
        (setSum, s) => setSum + (s.reps ?? 0) * (s.weight ?? 0),
        0,
      ),
    0,
  );

  const openNote = useCallback((note: string, exerciseName: string) => {
    setNoteExercise(exerciseName);
    setNote(note);
    setShowNote(true);
  }, []);

  const closeNote = useCallback(() => {
    setShowNote(false);
    setNote("");
    setNoteExercise("");
  }, []);

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <View style={styles.header}>
        <Text style={styles.title} accessibilityRole="header">
          Workout Complete
        </Text>
        {workout.completedAt && (
          <Text style={styles.subtitle}>
            {formatDuration(workout.startedAt, workout.completedAt)}
          </Text>
        )}
        {workout.tags?.length > 0 && (
          <View style={styles.headerTagList}>
            {workout.tags.map((tag) => (
              <View key={tag} style={styles.headerTagPill}>
                <Text style={styles.headerTagText}>{tag}</Text>
              </View>
            ))}
          </View>
        )}
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={[
          styles.scrollContent,
          { paddingBottom: insets.bottom + 24 },
        ]}
      >
        <View style={styles.statsRow}>
          <View style={styles.statCard}>
            <Text style={styles.statValue}>{workout.exercises.length}</Text>
            <Text style={styles.statLabel}>Exercises</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statValue}>{totalSets}</Text>
            <Text style={styles.statLabel}>Sets</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statValue}>{totalReps}</Text>
            <Text style={styles.statLabel}>Reps</Text>
          </View>
        </View>

        {totalVolume > 0 && (
          <View style={styles.volumeCard}>
            <Text style={styles.volumeLabel}>Total Volume</Text>
            <Text style={styles.volumeValue}>
              {totalVolume.toLocaleString()} kg
            </Text>
          </View>
        )}

        {workout.bodyWeight > 0 && (
          <View style={styles.volumeCard}>
            <Text style={styles.volumeLabel}>Body Weight</Text>
            <Text style={styles.volumeValue}>{workout.bodyWeight} kg</Text>
          </View>
        )}

        {workout.notes ? (
          <View style={styles.notesCard}>
            <Text style={styles.notesLabel}>Notes</Text>
            <Text style={styles.notesText}>{workout.notes}</Text>
          </View>
        ) : null}

        {workout.exercises.map((exercise) => {
          console.log(exercise);
          return(
          <View key={exercise.exerciseId} style={styles.exerciseSection}>
            <Text style={styles.exerciseName}>{exercise.name}</Text>
            {exercise?.note && (
              <Pressable
                onPress={() => openNote(exercise.note, exercise.name)}
                accessibilityRole="button"
                accessibilityLabel={`View note for ${exercise.name}`}
                style={styles.noteButton}
              >
                <DocumentAttachment size={16} />
                <Text style={styles.noteButtonText}>Saved note</Text>
              </Pressable>
            )}
            {exercise.sets.map((set) => (
              <View key={set.setId} style={styles.setRow}>
                <Text style={styles.setOrder}>Set {set.order}</Text>
                <Text style={styles.setDetail}>
                  {set.reps ?? '-'} reps
                  {set.weight !== null ? ` @ ${set.weight} ${set.weightUnit}` : ''}
                </Text>
              </View>
            ))}
          </View>
        )})}

        <Pressable
          onPress={handleDone}
          accessibilityRole="button"
          accessibilityLabel="Done"
          style={styles.doneButtonLarge}
        >
          <Text style={styles.doneButtonText}>Done</Text>
        </Pressable>
      </ScrollView>

      <Modal
        visible={showNote}
        transparent
        onRequestClose={closeNote}
      >
        <Pressable
          style={styles.popoverOverlay}
          onPress={closeNote}
        >
          <View style={styles.popoverContainer}>
            <View style={styles.popoverHeader}>
              <Text style={styles.popoverTitle}>Note - {noteExercise}</Text>
              <Pressable
                onPress={closeNote}
                accessibilityRole="button"
                accessibilityLabel="Close note"
                style={styles.popoverClose}
              >
                <Text style={styles.popoverCloseText}>Close</Text>
              </Pressable>
            </View>
            <Text style={styles.popoverText}>{note}</Text>
          </View>
        </Pressable>
      </Modal>
    </View>
  );
}

function formatDuration(startIso: string, endIso: string): string {
  const diffMs = new Date(endIso).getTime() - new Date(startIso).getTime();
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 60) return `${minutes} min`;
  const hours = Math.floor(minutes / 60);
  const remaining = minutes % 60;
  return `${hours}h ${remaining}m`;
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary.white,
  },
  header: {
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingVertical: 24,
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
  },
  subtitle: {
    fontSize: 14,
    color: colors.primary.greyDark,
    marginTop: 4,
  },
  headerTagList: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    marginTop: 10,
  },
  headerTagPill: {
    backgroundColor: colors.primary.blue,
    borderRadius: 12,
    paddingVertical: 4,
    paddingHorizontal: 10,
    marginHorizontal: 4,
    marginBottom: 4,
  },
  headerTagText: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.primary.white,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
  },
  emptyText: {
    fontSize: 16,
    color: colors.primary.greyDark,
    textAlign: 'center',
    marginTop: 48,
  },

  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  statCard: {
    flex: 1,
    alignItems: 'center',
    paddingVertical: 16,
    marginHorizontal: 4,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
  },
  statValue: {
    fontSize: 24,
    fontWeight: '700',
  },
  statLabel: {
    fontSize: 12,
    color: colors.primary.greyDark,
    marginTop: 4,
  },

  volumeCard: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    marginBottom: 16,
  },
  volumeLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.greyDark,
  },
  volumeValue: {
    fontSize: 18,
    fontWeight: '700',
  },

  notesCard: {
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    marginBottom: 16,
  },
  notesLabel: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.primary.greyDark,
    marginBottom: 4,
  },
  notesText: {
    fontSize: 14,
    color: colors.primary.greyDarkest,
  },

  exerciseSection: {
    marginBottom: 16,
  },
  exerciseName: {
    fontSize: 16,
    fontWeight: '700',
    marginBottom: 8,
  },
  setRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 6,
    paddingHorizontal: 12,
    backgroundColor: '#f9fafb',
    borderRadius: 6,
    marginBottom: 4,
  },
  setOrder: {
    fontSize: 14,
    color: colors.primary.greyDark,
    fontWeight: '600',
  },
  setDetail: {
    fontSize: 14,
  },

  doneButton: {
    marginTop: 24,
    alignSelf: 'center',
  },
  doneText: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.primary.blue,
  },
  doneButtonLarge: {
    backgroundColor: colors.primary.blue,
    paddingVertical: 14,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 24,
  },
  doneButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.primary.white,
  },
  noteButton: {
    paddingVertical: 8,
    paddingHorizontal: 12,
    marginBottom: 8,
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'flex-start',
    gap: '0.5rem',
    color: colors.primary.blue,
    width: 'fit-content',
  },
  noteButtonText: {
    fontSize: 14,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.4)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  modalContent: {
    width: '100%',
    maxHeight: '80%',
    backgroundColor: colors.primary.white,
    borderRadius: 12,
    padding: 16,
  },
  modalTitle: {
    fontSize: 16,
    fontWeight: '700',
    marginBottom: 8,
  },
  modalBody: {
    marginBottom: 12,
  },
  modalText: {
    fontSize: 14,
    color: colors.primary.greyDarkest,
  },
  modalCloseButton: {
    alignSelf: 'flex-end',
    paddingVertical: 8,
    paddingHorizontal: 12,
  },
  modalCloseText: {
    color: colors.primary.blue,
    fontWeight: '600',
    fontSize: 14,
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
