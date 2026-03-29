import { Modal, Pressable, View, StyleSheet, Text, Platform } from "react-native";
import { WorkoutExercise } from "../types/workout";
import { colors } from "../constants/colors";

type PeviousExerciseDataProps = {
  previousExercise?: WorkoutExercise;
  visible: boolean;
  onClose: () => void;
}

export const PreviousExerciseData = ({ previousExercise, visible, onClose }: PeviousExerciseDataProps) => {
  if (!previousExercise) { return null };
  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
        <Pressable
          style={styles.popoverOverlay}
          onPress={onClose}
        >
          <View style={styles.popoverContainer}>
            <View style={styles.popoverHeader}>
              <Text style={styles.popoverTitle}>{previousExercise.name}</Text>
              <Pressable
                onPress={onClose}
                accessibilityRole="button"
                accessibilityLabel="Close note"
                style={styles.popoverClose}
              >
                <Text style={styles.popoverCloseText}>Close</Text>
              </Pressable>
            </View>
            <View style={styles.setContainer}>
              {previousExercise?.note && (
                <View style={styles.notesCard}>
                  <Text style={styles.notesLabel}>Notes</Text>
                  <Text style={styles.notesText}>{previousExercise.note}</Text>
                </View>
              )}
              <Text>Previous set(s) data</Text>
              {previousExercise.sets.map((setData) => (
                <Text key={setData.setId} >
                  {setData.order}: {setData.reps} reps @ {setData.weight}{setData.weightUnit}
                </Text>
              ))}
          </View>
        </View>
      </Pressable>
    </Modal>
  );
};

const styles = StyleSheet.create({
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

  setContainer: {
    display: 'flex',
    gap: '0.5rem',
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

});
