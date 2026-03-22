import React, { useCallback, useMemo, useState } from 'react';
import {
  View,
  Text,
  TextInput,
  FlatList,
  Pressable,
  StyleSheet,
  Platform,
  Switch,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useExercise } from '../contexts/ExerciseContext';
import { Button, ConfirmationDialog } from '../components';
import { ExerciseEditorModal } from '../components/ExerciseEditorModal';
import { colors } from '../constants/colors';
import type { SavedExercise, SavedExercisesScreenProps } from '../types/workout';
import { getPublicInstanceFromRootTag } from 'react-native/types_generated/Libraries/ReactNative/RendererImplementation';

export function SavedExercisesScreen({ navigation }: SavedExercisesScreenProps) {
  const insets = useSafeAreaInsets();
  const {
    allSavedExercises,
    saveExercise,
    updateExercise,
    archiveExercise,
    restoreExercise,
  } = useExercise();

  const [searchQuery, setSearchQuery] = useState('');
  const [newExerciseName, setNewExerciseName] = useState('');
  const [editorExercise, setEditorExercise] = useState<SavedExercise | null>(null);
  const [archiveTarget, setArchiveTarget] = useState<SavedExercise | null>(null);
  const [showArchived, setShowArchived] = useState(false);

  const filteredExercises = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    let exercises = allSavedExercises;

    // Filter by archived status
    if (!showArchived) {
      exercises = exercises.filter((e) => {
        console.log(e, !e.archivedAt, e.archivedAt === null, !e?.archivedAt || e.archivedAt === null);
        return !e?.archivedAt || e.archivedAt === null;
      })
    }

    // Filter by search query
    if (query) {
      exercises = exercises.filter(
        (e) =>
          e.name.toLowerCase().includes(query) ||
          e.tags.some((t) => t.toLowerCase().includes(query)),
      );
    }

    return exercises;
  }, [allSavedExercises, searchQuery, showArchived]);

  const archivedCount = useMemo(
    () => allSavedExercises.filter((e) => e.archivedAt && e.archivedAt !== null).length,
    [allSavedExercises],
  );

  const handleAddExercise = useCallback(() => {
    const trimmed = newExerciseName.trim();
    if (!trimmed) return;
    const exercise = saveExercise(trimmed);
    setNewExerciseName('');
    setEditorExercise(exercise);
  }, [newExerciseName, saveExercise]);

  const handleEditorSave = useCallback(
    async (updated: SavedExercise) => {
      await updateExercise(updated);
      setEditorExercise(null);
    },
    [updateExercise],
  );

  const handleConfirmArchive = useCallback(async () => {
    if (!archiveTarget) return;
    await archiveExercise(archiveTarget.savedExerciseId);
    setArchiveTarget(null);
  }, [archiveTarget, archiveExercise]);

  const handleRestore = useCallback(
    async (savedExerciseId: string) => {
      await restoreExercise(savedExerciseId);
    },
    [restoreExercise],
  );

  const renderExerciseItem = useCallback(
    ({ item }: { item: SavedExercise }) => {
      const isArchived = item.archivedAt && item.archivedAt !== null;
      return (
        <View style={[styles.exerciseItem, isArchived && styles.exerciseItemArchived]}>
          <View style={styles.exerciseItemContent}>
            <View style={styles.exerciseNameRow}>
              <Text style={[styles.exerciseItemName, isArchived && styles.exerciseItemNameArchived]}>
                {item.name}
              </Text>
              {isArchived && (
                <View style={styles.archivedBadge}>
                  <Text style={styles.archivedBadgeText}>Archived</Text>
                </View>
              )}
            </View>
            {item.tags.length > 0 && (
              <View style={styles.tagRow}>
                {item.tags.map((tag) => (
                  <View key={tag} style={[styles.tagPill, isArchived && styles.tagPillArchived]}>
                    <Text style={styles.tagPillText}>{tag}</Text>
                  </View>
                ))}
              </View>
            )}
            {item.note ? (
              <Text
                style={[styles.notePreview, isArchived && styles.notePreviewArchived]}
                numberOfLines={2}
              >
                {item.note}
              </Text>
            ) : null}
          </View>
          <View style={styles.exerciseItemActions}>
            {isArchived ? (
              <Pressable
                onPress={() => handleRestore(item.savedExerciseId)}
                accessibilityRole="button"
                accessibilityLabel={`Restore ${item.name}`}
                style={styles.restoreButton}
              >
                <Text style={styles.restoreButtonText}>Restore</Text>
              </Pressable>
            ) : (
              <>
                <Pressable
                  onPress={() => setEditorExercise(item)}
                  accessibilityRole="button"
                  accessibilityLabel={`Edit ${item.name}`}
                  style={styles.editButton}
                >
                  <Text style={styles.editButtonText}>Edit</Text>
                </Pressable>
                <Pressable
                  onPress={() => setArchiveTarget(item)}
                  accessibilityRole="button"
                  accessibilityLabel={`Archive ${item.name}`}
                  style={styles.archiveButton}
                >
                  <Text style={styles.archiveButtonText}>Archive</Text>
                </Pressable>
              </>
            )}
          </View>
        </View>
      );
    },
    [handleRestore],
  );

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <View style={styles.header}>
        <Pressable
          onPress={() => navigation.goBack()}
          accessibilityRole="button"
          accessibilityLabel="Go back"
          style={styles.backButton}
        >
          <Text style={styles.backText}>Back</Text>
        </Pressable>
        <Text style={styles.title} accessibilityRole="header">
          Saved Exercises
        </Text>
        <View style={styles.backButton} />
      </View>

      <View style={styles.content}>
        <TextInput
          style={[
            styles.searchInput,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          placeholder="Search by name or tag..."
          placeholderTextColor={colors.primary.grey}
          value={searchQuery}
          onChangeText={setSearchQuery}
          returnKeyType="search"
          accessibilityLabel="Search exercises"
        />

        <View style={styles.addRow}>
          <TextInput
            style={[
              styles.addInput,
              Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
            ]}
            placeholder="New exercise name"
            placeholderTextColor={colors.primary.grey}
            value={newExerciseName}
            onChangeText={setNewExerciseName}
            onSubmitEditing={handleAddExercise}
            returnKeyType="done"
            accessibilityLabel="New exercise name"
          />
          <Button
            title="Add"
            variant="primary"
            onPress={handleAddExercise}
            disabled={!newExerciseName.trim()}
            containerStyle={styles.addButton}
          />
        </View>

        {archivedCount > 0 && (
          <View style={styles.archivedToggleRow}>
            <Text style={styles.archivedToggleLabel}>
              Show archived ({archivedCount})
            </Text>
            <Switch
              value={showArchived}
              onValueChange={setShowArchived}
              trackColor={{ false: colors.primary.greyLight, true: colors.primary.blue }}
              accessibilityLabel="Show archived exercises"
            />
          </View>
        )}

        {filteredExercises.length === 0 ? (
          <View style={styles.emptyState}>
            <Text style={styles.emptyText}>
              {allSavedExercises.filter((e) => e.archivedAt === null).length === 0
                ? 'No saved exercises yet. Add one above or save exercises during a workout.'
                : 'No exercises match your search.'}
            </Text>
          </View>
        ) : (
          <FlatList
            data={filteredExercises}
            keyExtractor={(item) => item.savedExerciseId}
            renderItem={renderExerciseItem}
            contentContainerStyle={{ paddingBottom: insets.bottom + 16 }}
            keyboardShouldPersistTaps="handled"
          />
        )}
      </View>

      {editorExercise && (
        <ExerciseEditorModal
          visible={!!editorExercise}
          exercise={editorExercise}
          onSave={handleEditorSave}
          onCancel={() => setEditorExercise(null)}
        />
      )}

      <ConfirmationDialog
        visible={archiveTarget !== null}
        title="Archive exercise?"
        message={`"${archiveTarget?.name ?? ''}" will be archived. You can restore it later from the archived list.`}
        confirmLabel="Archive"
        cancelLabel="Cancel"
        destructive
        onConfirm={handleConfirmArchive}
        onCancel={() => setArchiveTarget(null)}
      />
    </View>
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
    paddingHorizontal: 24,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  backButton: {
    minWidth: 50,
    paddingVertical: 4,
  },
  backText: {
    fontSize: 16,
    color: colors.primary.blue,
    fontWeight: '600',
  },
  title: {
    fontSize: 20,
    fontWeight: '700',
    textAlign: 'center',
    flex: 1,
  },
  content: {
    flex: 1,
    padding: 24,
  },
  searchInput: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    marginBottom: 16,
  },
  addRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  addInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    marginRight: 8,
  },
  addButton: {
    minHeight: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
  },
  archivedToggleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
    paddingVertical: 8,
    paddingHorizontal: 4,
  },
  archivedToggleLabel: {
    fontSize: 14,
    color: colors.primary.greyDark,
    fontWeight: '500',
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  emptyText: {
    fontSize: 15,
    color: colors.primary.greyDark,
    textAlign: 'center',
    lineHeight: 22,
  },
  exerciseItem: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
  },
  exerciseItemArchived: {
    opacity: 0.5,
  },
  exerciseItemContent: {
    flex: 1,
    marginRight: 12,
  },
  exerciseNameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 4,
  },
  exerciseItemName: {
    fontSize: 16,
    fontWeight: '700',
  },
  exerciseItemNameArchived: {
    color: colors.primary.greyDark,
  },
  archivedBadge: {
    backgroundColor: colors.primary.greyLight,
    borderRadius: 10,
    paddingVertical: 2,
    paddingHorizontal: 8,
  },
  archivedBadgeText: {
    fontSize: 11,
    fontWeight: '600',
    color: colors.primary.greyDark,
  },
  tagRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 4,
    marginBottom: 4,
  },
  tagPill: {
    backgroundColor: colors.primary.blue,
    borderRadius: 12,
    paddingVertical: 3,
    paddingHorizontal: 10,
    marginRight: 6,
    marginBottom: 4,
  },
  tagPillArchived: {
    backgroundColor: colors.primary.grey,
  },
  tagPillText: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.primary.white,
  },
  notePreview: {
    fontSize: 13,
    color: colors.primary.greyDark,
    marginTop: 4,
    lineHeight: 18,
  },
  notePreviewArchived: {
    color: colors.primary.grey,
  },
  exerciseItemActions: {
    alignItems: 'flex-end',
    gap: 8,
  },
  editButton: {
    paddingVertical: 6,
    paddingHorizontal: 12,
  },
  editButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.blue,
  },
  archiveButton: {
    paddingVertical: 6,
    paddingHorizontal: 12,
  },
  archiveButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.red,
  },
  restoreButton: {
    paddingVertical: 6,
    paddingHorizontal: 12,
  },
  restoreButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.primary.blue,
  },
});
