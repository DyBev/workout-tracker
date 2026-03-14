import { useState, useCallback } from 'react';
import { View, Text, TextInput, Pressable, Platform, StyleSheet } from 'react-native';
import { Button } from './Button';
import { colors } from '../constants/colors';

interface TagInputProps {
  tags: string[];
  onUpdateTags: (tags: string[]) => void;
}

export function TagInput({ tags, onUpdateTags }: TagInputProps) {
  const [value, setValue] = useState('');

  const handleAdd = useCallback(() => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (tags.includes(trimmed)) {
      setValue('');
      return;
    }
    onUpdateTags([...tags, trimmed]);
    setValue('');
  }, [value, tags, onUpdateTags]);

  const handleRemove = useCallback(
    (tag: string) => {
      onUpdateTags(tags.filter((t) => t !== tag));
    },
    [tags, onUpdateTags],
  );

  return (
    <View style={styles.tagSection}>
      <Text style={styles.tagLabel}>Tags</Text>
      {tags.length > 0 && (
        <View style={styles.tagList}>
          {tags.map((tag) => (
            <Pressable
              key={tag}
              onPress={() => handleRemove(tag)}
              accessibilityRole="button"
              accessibilityLabel={`Remove tag ${tag}`}
              style={styles.tagPill}
            >
              <Text style={styles.tagPillText}>{tag}</Text>
              <Text style={styles.tagPillRemove}>x</Text>
            </Pressable>
          ))}
        </View>
      )}
      <View style={styles.tagInputRow}>
        <TextInput
          style={[
            styles.tagInput,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          placeholder="e.g. Push, Pull, Legs"
          placeholderTextColor={colors.primary.grey}
          value={value}
          onChangeText={setValue}
          onSubmitEditing={handleAdd}
          returnKeyType="done"
          autoCapitalize="none"
          accessibilityLabel="Add tag"
        />
        <Button
          title="Add"
          variant="secondary"
          onPress={handleAdd}
          disabled={!value.trim()}
          containerStyle={styles.tagAddButton}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({

  tagSection: {
    marginBottom: 24,
  },
  tagLabel: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
  },
  tagList: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 8,
  },
  tagPill: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.primary.blue,
    borderRadius: 16,
    paddingVertical: 6,
    paddingHorizontal: 12,
    marginRight: 8,
    marginBottom: 8,
  },
  tagPillText: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.primary.white,
  },
  tagPillRemove: {
    fontSize: 13,
    fontWeight: '700',
    color: colors.primary.white,
    marginLeft: 6,
    opacity: 0.8,
  },
  tagInputRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  tagInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    marginRight: 8,
  },
  tagAddButton: {
    minHeight: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
  },
});
