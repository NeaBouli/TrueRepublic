import { type InputHTMLAttributes, type ReactNode, useId } from 'react';
import clsx from 'clsx';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: ReactNode;
}

export function Input({
  label,
  error,
  helperText,
  className,
  ...props
}: InputProps) {
  const generatedId = useId();
  const inputId = props.id ?? generatedId;
  const descriptionId = `${inputId}-description`;
  const describedBy =
    [props['aria-describedby'], (error || helperText) && descriptionId]
      .filter(Boolean)
      .join(' ') || undefined;

  return (
    <div className="w-full">
      {label && (
        <label
          htmlFor={inputId}
          className="block text-sm font-medium text-gray-700 mb-1"
        >
          {label}
        </label>
      )}
      <input
        {...props}
        id={inputId}
        aria-describedby={describedBy}
        aria-invalid={error ? true : props['aria-invalid']}
        className={clsx('input', error && 'border-red-500', className)}
      />
      {error && (
        <p id={descriptionId} className="mt-1 text-sm text-red-600">
          {error}
        </p>
      )}
      {helperText && !error && (
        <p id={descriptionId} className="mt-1 text-sm text-gray-500">
          {helperText}
        </p>
      )}
    </div>
  );
}
