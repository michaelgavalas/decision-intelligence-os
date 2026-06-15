import type { UserError } from "./auth-context";

/** Friendly copy for known backend error codes. */
const CODE_COPY: Record<string, string> = {
  INVALID_CREDENTIALS: "Incorrect email or password.",
  RATE_LIMITED: "Too many attempts. Please try again shortly.",
  EMAIL_TAKEN: "An account with this email already exists.",
  WEAK_PASSWORD: "Please choose a stronger password (at least 8 characters).",
  INVALID_EMAIL: "Please enter a valid email address.",
  INVALID_NAME: "Please enter your name.",
  NETWORK_ERROR: "Something went wrong. Please try again.",
};

/** Maps a user error code to friendly copy, falling back to its message. */
export function friendlyMessage(error: UserError): string {
  return CODE_COPY[error.code] ?? error.message;
}

/** The form fields that backend errors may be attached to. */
export type FieldName = "email" | "name" | "password";

/**
 * Splits user errors into per-field messages and remaining form-level messages.
 * INVALID_CREDENTIALS is treated as form-level because it intentionally avoids
 * revealing which field was wrong.
 */
export function mapUserErrors(errors: UserError[]): {
  fieldErrors: Partial<Record<FieldName, string>>;
  formErrors: string[];
} {
  const fieldErrors: Partial<Record<FieldName, string>> = {};
  const formErrors: string[] = [];

  for (const error of errors) {
    const message = friendlyMessage(error);
    const field = error.field;

    if (
      error.code !== "INVALID_CREDENTIALS" &&
      (field === "email" || field === "name" || field === "password")
    ) {
      if (!fieldErrors[field]) {
        fieldErrors[field] = message;
      }
    } else {
      formErrors.push(message);
    }
  }

  return { fieldErrors, formErrors };
}
