import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  Pressable,
  ScrollView,
  StyleSheet,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useWorkout } from '../contexts/WorkoutContext';
import { colors } from '../constants/colors';
import type { WorkoutHistoryScreenProps } from '../types/workout';
import type { Workout } from '../types/workout';

interface WorkoutRowProps {
  workout: Workout;
  onPress: (workoutId: string) => void;
}

function WorkoutRow({ workout, onPress }: WorkoutRowProps) {
  const handlePress = useCallback(() => {
    onPress(workout.workoutId);
  }, [workout.workoutId, onPress]);

  const exerciseCount = workout.exercises.length;
  const totalSets = workout.exercises.reduce(
    (sum, ex) => sum + ex.sets.length,
    0,
  );

  return (
    <Pressable
      style={styles.row}
      onPress={handlePress}
      accessibilityRole="button"
      accessibilityLabel={`Workout on ${formatDate(workout.startedAt)}, ${exerciseCount} exercises, ${totalSets} sets`}
    >
      <View style={styles.rowLeft}>
        <Text style={styles.rowDate}>{formatDate(workout.startedAt)}</Text>
        <Text style={styles.rowTime}>{formatTime(workout.startedAt)}</Text>
        {workout.tags?.length > 0 && (
          <View style={styles.rowTagList}>
            {workout.tags.map((tag) => (
              <View key={tag} style={styles.rowTagPill}>
                <Text style={styles.rowTagText}>{tag}</Text>
              </View>
            ))}
          </View>
        )}
      </View>
      <View style={styles.rowRight}>
        <Text style={styles.rowStat}>
          {exerciseCount} exercise{exerciseCount !== 1 ? 's' : ''}
        </Text>
        <Text style={styles.rowStat}>
          {totalSets} set{totalSets !== 1 ? 's' : ''}
        </Text>
        {workout.completedAt && (
          <Text style={styles.rowDuration}>
            {formatDuration(workout.startedAt, workout.completedAt)}
          </Text>
        )}
      </View>
    </Pressable>
  );
}

export function WorkoutHistoryScreen({
  navigation,
}: WorkoutHistoryScreenProps) {
  const { state, loadHistory } = useWorkout();
  const insets = useSafeAreaInsets();
  const [activeFilters, setActiveFilters] = useState<string[]>([]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  const allTags = useMemo(() => {
    const tagSet = new Set<string>();
    for (const workout of state.history) {
      if (!workout.tags) continue;
      for (const tag of workout.tags) {
        tagSet.add(tag);
      }
    }
    return Array.from(tagSet).sort();
  }, [state.history]);

  const filteredHistory = useMemo(() => {
    if (activeFilters.length === 0) return state.history;
    return state.history.filter((workout) =>
      activeFilters.every((filter) => workout.tags.includes(filter)),
    );
  }, [state.history, activeFilters]);

  const toggleFilter = useCallback((tag: string) => {
    setActiveFilters((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag],
    );
  }, []);

  const clearFilters = useCallback(() => {
    setActiveFilters([]);
  }, []);

  const handlePressWorkout = useCallback(
    (workoutId: string) => {
      navigation.navigate('WorkoutSummary', { workoutId });
    },
    [navigation],
  );

  const handleGoHome = useCallback(() => {
    navigation.navigate('Home');
  }, [navigation]);

  const renderItem = useCallback(
    ({ item }: { item: Workout }) => (
      <WorkoutRow workout={item} onPress={handlePressWorkout} />
    ),
    [handlePressWorkout],
  );

  const keyExtractor = useCallback((item: Workout) => item.workoutId, []);

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <View style={styles.header}>
        <Pressable
          onPress={handleGoHome}
          accessibilityRole="button"
          accessibilityLabel="Go back"
          style={styles.backButton}
        >
          <Text style={styles.backText}>Home</Text>
        </Pressable>
        <Text style={styles.title} accessibilityRole="header">
          History
        </Text>
        <View style={styles.backButton} />
      </View>

      {allTags.length > 0 && (
        <View style={styles.filterBar}>
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={styles.filterScrollContent}
          >
            {activeFilters.length > 0 && (
              <Pressable
                onPress={clearFilters}
                accessibilityRole="button"
                accessibilityLabel="Clear all filters"
                style={styles.clearFilterPill}
              >
                <Text style={styles.clearFilterText}>Clear</Text>
              </Pressable>
            )}
            {allTags.map((tag) => {
              const isActive = activeFilters.includes(tag);
              return (
                <Pressable
                  key={tag}
                  onPress={() => toggleFilter(tag)}
                  accessibilityRole="button"
                  accessibilityLabel={`${isActive ? 'Remove' : 'Add'} filter ${tag}`}
                  accessibilityState={{ selected: isActive }}
                  style={[
                    styles.filterPill,
                    isActive && styles.filterPillActive,
                  ]}
                >
                  <Text
                    style={[
                      styles.filterPillText,
                      isActive && styles.filterPillTextActive,
                    ]}
                  >
                    {tag}
                  </Text>
                </Pressable>
              );
            })}
          </ScrollView>
        </View>
      )}

      {filteredHistory.length === 0 ? (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>
            {state.history.length === 0
              ? 'No workouts yet. Start your first workout!'
              : 'No workouts match the selected filters.'}
          </Text>
        </View>
      ) : (
        <FlatList
          data={filteredHistory}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          contentContainerStyle={[
            styles.listContent,
            { paddingBottom: insets.bottom + 16 },
          ]}
        />
      )}
    </View>
  );
}

function formatDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

function formatTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatDuration(startIso: string, endIso: string): string {
  const diffMs = new Date(endIso).getTime() - new Date(startIso).getTime();
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 60) return `${minutes}m`;
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
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  backButton: {
    width: 60,
  },
  backText: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.primary.blue,
  },
  title: {
    fontSize: 20,
    fontWeight: '700',
    textAlign: 'center',
    flex: 1,
  },

  filterBar: {
    borderBottomWidth: 1,
    borderBottomColor: colors.primary.greyLight,
  },
  filterScrollContent: {
    paddingHorizontal: 24,
    paddingVertical: 12,
    gap: 8,
    flexDirection: 'row',
    alignItems: 'center',
  },
  filterPill: {
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 16,
    paddingVertical: 6,
    paddingHorizontal: 14,
  },
  filterPillActive: {
    backgroundColor: colors.primary.blue,
    borderColor: colors.primary.blue,
  },
  filterPillText: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.primary.greyDarkest,
  },
  filterPillTextActive: {
    color: colors.primary.white,
  },
  clearFilterPill: {
    borderWidth: 1,
    borderColor: colors.primary.red,
    borderRadius: 16,
    paddingVertical: 6,
    paddingHorizontal: 14,
  },
  clearFilterText: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.primary.red,
  },

  listContent: {
    padding: 24,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    marginBottom: 12,
  },
  rowLeft: {
    flex: 1,
  },
  rowDate: {
    fontSize: 15,
    fontWeight: '600',
  },
  rowTime: {
    fontSize: 13,
    color: colors.primary.greyDark,
    marginTop: 2,
  },
  rowRight: {
    alignItems: 'flex-end',
  },
  rowStat: {
    fontSize: 13,
    color: colors.primary.greyDark,
  },
  rowDuration: {
    fontSize: 13,
    color: colors.primary.greyDarkest,
    fontWeight: '600',
    marginTop: 2,
  },
  rowTagList: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 6,
  },
  rowTagPill: {
    backgroundColor: colors.primary.blue,
    borderRadius: 10,
    paddingVertical: 2,
    paddingHorizontal: 8,
    marginRight: 6,
    marginBottom: 4,
  },
  rowTagText: {
    fontSize: 11,
    fontWeight: '600',
    color: colors.primary.white,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 48,
  },
  emptyText: {
    fontSize: 16,
    color: colors.primary.greyDark,
    textAlign: 'center',
  },
});
