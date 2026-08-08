import { type TextareaHTMLAttributes, useId } from 'react';
import clsx from 'clsx';

interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export function TextArea({
  label,
  error,
  helperText,
  className,
  ...props
}: TextAreaProps) {
  const generatedId = useId();
  const textAreaId = props.id ?? generatedId;
  const descriptionId = `${textAreaId}-description`;
  const describedBy =
    [props['aria-describedby'], (error || helperText) && descriptionId]
      .filter(Boolean)
      .join(' ') || undefined;

  return (
    <div className="w-full">
      {label && (
        <label
          htmlFor={textAreaId}
          className="block text-sm font-medium text-gray-700 mb-1"
        >
          {label}
        </label>
      )}
      <textarea
        {...props}
        id={textAreaId}
        aria-describedby={describedBy}
        aria-invalid={error ? true : props['aria-invalid']}
        className={clsx(
          'input resize-none',
          error && 'border-red-500',
          className
        )}
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
