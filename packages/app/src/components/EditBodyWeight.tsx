import { Modal, StyleSheet, View, Text } from 'react-native';
import { Button } from './Button';
import { colors } from '../constants/colors';
import { BodyWeightInput } from './BodyWeightInput';
import { useState } from 'react';
import { updateBodyWeight } from '../services/workoutApi';

type EditBodyWeightProps = {
  bodyWeightValue: number,
  workoutId: string,
  visible: boolean,
  onClose: () => void,
  updateWorkoutBodyWeight: (_: number) => void,
}

export const EditBodyWeight = ({
  bodyWeightValue,
  workoutId,
  visible,
  onClose,
  updateWorkoutBodyWeight,
}: EditBodyWeightProps) => {
  const [bodyWeight, setBodyWeight] = useState(bodyWeightValue);

  const onChange = (bodyWeight: number) => {
    setBodyWeight(bodyWeight);
  };

  const onSave = () => {
    updateBodyWeight(bodyWeight, workoutId).then((success) => {
      if (success) {
        updateWorkoutBodyWeight(bodyWeight);
        onClose();
      }
    });
  }

  return(
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
      statusBarTranslucent
    >
      <View style={styles.overlay}>
        <View
          style={styles.dialog}
          accessibilityRole="alert"
          accessibilityLabel={"Update body weight"}
        >
          <Text style={styles.title}>Update Body Weight</Text>
          <BodyWeightInput
            value={bodyWeight}
            onChangeValue={onChange}
          />
          <View style={styles.actions}>
            <Button
              title={"cancel"}
              variant='secondary'
              onPress={onClose}
              containerStyle={styles.button}
            />
            <Button
              title={"save"}
              variant='primary'
              onPress={onSave}
              containerStyle={StyleSheet.flatten([
                styles.button,
              ])}
            />
          </View>
        </View>
      </View>
    </Modal>
  )
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  dialog: {
    backgroundColor: colors.primary.white,
    borderRadius: 12,
    padding: 24,
    width: '100%',
    maxWidth: 340,
  },
  title: {
    fontSize: 18,
    fontWeight: '700',
    color: colors.primary.black,
    marginBottom: 8,
  },
  message: {
    fontSize: 14,
    color: colors.primary.greyDark,
    lineHeight: 20,
    marginBottom: 24,
  },
  actions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 12,
  },
  button: {
    flex: 1,
    minHeight: 44,
    paddingVertical: 10,
  },
  destructiveButton: {
    backgroundColor: colors.primary.red,
  },
});
