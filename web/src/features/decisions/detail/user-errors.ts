/** A single field-level error returned by a GraphQL mutation payload. */
export interface UserErrorLike {
  field?: string | null;
  message: string;
  code?: string;
}

export interface MappedUserErrors {
  /** Errors keyed by their `field`, for inline display next to inputs. */
  fields: Record<string, string>;
  /** The first error without a field, suitable for a form-level alert. */
  formError: string | null;
}

/**
 * Splits a list of `userErrors` into per-field messages and a single
 * form-level message for errors that are not tied to a specific field.
 */
export function userErrorsToMap(
  errors: ReadonlyArray<UserErrorLike>,
): MappedUserErrors {
  const fields: Record<string, string> = {};
  let formError: string | null = null;

  for (const error of errors) {
    if (error.field) {
      fields[error.field] = error.message;
    } else if (!formError) {
      formError = error.message;
    }
  }

  // Fall back to the first message so nothing is silently swallowed.
  if (!formError && errors.length > 0 && Object.keys(fields).length === 0) {
    formError = errors[0].message;
  }

  return { fields, formError };
}
