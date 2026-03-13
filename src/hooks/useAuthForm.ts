import { useCallback, useRef, useState } from 'react';
import type { TextInput } from 'react-native';

interface UseAuthFormOptions<T extends Record<string, string>> {
  initialValues: T;
  validate: (values: T) => Partial<Record<keyof T, string>>;
  onSubmit: (values: T) => Promise<void>;
}

interface UseAuthFormReturn<T extends Record<string, string>> {
  values: T;
  errors: Partial<Record<keyof T, string>>;
  formError: string | null;
  isSubmitting: boolean;
  setValue: (field: keyof T, value: string) => void;
  setFormError: (error: string | null) => void;
  handleSubmit: () => Promise<void>;
  resetForm: () => void;
  createRef: () => React.RefObject<TextInput | null>;
  refs: React.MutableRefObject<React.RefObject<TextInput | null>[]>;
}

export function useAuthForm<T extends Record<string, string>>({
  initialValues,
  validate,
  onSubmit,
}: UseAuthFormOptions<T>): UseAuthFormReturn<T> {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const refs = useRef<React.RefObject<TextInput | null>[]>([]);

  const setValue = useCallback((field: keyof T, value: string) => {
    setValues((prev) => ({ ...prev, [field]: value }));
    // Clear field error when user starts typing
    setErrors((prev) => {
      if (prev[field]) {
        const next = { ...prev };
        delete next[field];
        return next;
      }
      return prev;
    });
    setFormError(null);
  }, []);

  const handleSubmit = useCallback(async () => {
    setFormError(null);

    const validationErrors = validate(values);
    const hasErrors = Object.keys(validationErrors).length > 0;

    if (hasErrors) {
      setErrors(validationErrors);
      return;
    }

    setErrors({});
    setIsSubmitting(true);

    try {
      await onSubmit(values);
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'An unexpected error occurred';
      setFormError(message);
    } finally {
      setIsSubmitting(false);
    }
  }, [values, validate, onSubmit]);

  const resetForm = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setFormError(null);
    setIsSubmitting(false);
  }, [initialValues]);

  const createRef = useCallback(() => {
    const ref: React.RefObject<TextInput | null> = { current: null };
    refs.current.push(ref);
    return ref;
  }, []);

  return {
    values,
    errors,
    formError,
    isSubmitting,
    setValue,
    setFormError,
    handleSubmit,
    resetForm,
    createRef,
    refs,
  };
}
