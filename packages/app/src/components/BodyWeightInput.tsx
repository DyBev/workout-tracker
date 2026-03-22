import { useCallback, useEffect, useRef, useState } from 'react';
import { View, Text, TextInput, Platform, StyleSheet } from 'react-native';
import { colors } from '../constants/colors';

interface BodyWeightInputProps {
  value: number;
  onChangeValue: (value: number) => void;
}

export function BodyWeightInput({ value, onChangeValue }: BodyWeightInputProps) {
  const [input, setInput] = useState<string>(
    value ? String(value) : '',
  );
  const isFocused = useRef(false);

  useEffect(() => {
    if (!isFocused.current) {
      setInput(value ? String(value) : '');
    }
  }, [value]);

  const handleChange = useCallback(
    (text: string) => {
      if (text !== '' && isNaN(parseFloat(text)) && text !== '.') {
        return;
      }
      setInput(text);

      if (text === '' || text === '.') {
        onChangeValue(text === '' ? 0 : value);
      } else {
        const parsed = parseFloat(text);
        if (!isNaN(parsed)) {
          onChangeValue(parsed);
        }
      }
    },
    [value, onChangeValue],
  );

  const handleBlur = useCallback(() => {
    isFocused.current = false;
    const parsed = parseFloat(input);
    if (input === '' || isNaN(parsed)) {
      setInput('');
      onChangeValue(0);
    } else {
      setInput(String(parsed));
      onChangeValue(parsed);
    }
  }, [input, onChangeValue]);

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Body Weight</Text>
      <View style={styles.inputRow}>
        <TextInput
          style={[
            styles.input,
            Platform.OS === 'web' && { outlineStyle: 'none' as unknown as undefined },
          ]}
          placeholder="Optional"
          placeholderTextColor={colors.primary.grey}
          keyboardType="decimal-pad"
          value={input}
          onChangeText={handleChange}
          onFocus={() => { isFocused.current = true; }}
          onBlur={handleBlur}
          accessibilityLabel="Body weight in kilograms"
        />
        <Text style={styles.unitLabel}>kg</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
  },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  input: {
    flex: 1,
    borderWidth: 1,
    borderColor: colors.primary.greyLight,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 14,
    minHeight: 48,
  },
  unitLabel: {
    fontSize: 13,
    color: colors.primary.greyDark,
    marginLeft: 8,
  },
});
