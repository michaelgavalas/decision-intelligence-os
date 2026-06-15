import { useRef, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { mapUserErrors, type FieldName } from "@/features/auth/error-copy";
import { useAuth } from "@/features/auth/use-auth";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MIN_PASSWORD_LENGTH = 8;

interface LocationState {
  from?: string;
}

export function RegisterPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as LocationState | null)?.from ?? "/decisions";

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<
    Partial<Record<FieldName, string>>
  >({});
  const [formErrors, setFormErrors] = useState<string[]>([]);
  const [pending, setPending] = useState(false);

  const nameRef = useRef<HTMLInputElement>(null);
  const emailRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  function validate(): Partial<Record<FieldName, string>> {
    const errors: Partial<Record<FieldName, string>> = {};
    if (!name.trim()) {
      errors.name = "Name is required.";
    }
    if (!email.trim()) {
      errors.email = "Email is required.";
    } else if (!EMAIL_PATTERN.test(email.trim())) {
      errors.email = "Please enter a valid email address.";
    }
    if (!password) {
      errors.password = "Password is required.";
    } else if (password.length < MIN_PASSWORD_LENGTH) {
      errors.password = `Password must be at least ${MIN_PASSWORD_LENGTH} characters.`;
    }
    return errors;
  }

  function focusFirstInvalid(errors: Partial<Record<FieldName, string>>) {
    if (errors.name) {
      nameRef.current?.focus();
    } else if (errors.email) {
      emailRef.current?.focus();
    } else if (errors.password) {
      passwordRef.current?.focus();
    }
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormErrors([]);

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setFieldErrors(validationErrors);
      focusFirstInvalid(validationErrors);
      return;
    }
    setFieldErrors({});

    setPending(true);
    const result = await auth.register({
      name: name.trim(),
      email: email.trim(),
      password,
    });
    setPending(false);

    if (result.ok) {
      navigate(from, { replace: true });
      return;
    }

    const { fieldErrors: mappedFields, formErrors: mappedForm } = mapUserErrors(
      result.userErrors,
    );
    setFieldErrors(mappedFields);
    setFormErrors(mappedForm);
    focusFirstInvalid(mappedFields);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create your account</CardTitle>
        <CardDescription>
          Start tracking decision quality with your team.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
          {formErrors.length > 0 ? (
            <div
              role="alert"
              aria-live="assertive"
              className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
            >
              {formErrors.map((message) => (
                <p key={message}>{message}</p>
              ))}
            </div>
          ) : null}

          <FormField label="Name" error={fieldErrors.name} required>
            {(props) => (
              <Input
                {...props}
                ref={nameRef}
                autoComplete="name"
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            )}
          </FormField>

          <FormField label="Email" error={fieldErrors.email} required>
            {(props) => (
              <Input
                {...props}
                ref={emailRef}
                type="email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
              />
            )}
          </FormField>

          <FormField
            label="Password"
            error={fieldErrors.password}
            helperText={`At least ${MIN_PASSWORD_LENGTH} characters.`}
            required
          >
            {(props) => (
              <PasswordInput
                {...props}
                ref={passwordRef}
                autoComplete="new-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            )}
          </FormField>

          <Button type="submit" loading={pending} disabled={pending}>
            Create account
          </Button>
        </form>

        <p className="mt-4 text-sm text-muted-foreground">
          Have an account?{" "}
          <Link
            to="/login"
            className="font-medium text-primary hover:underline"
          >
            Sign in
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
