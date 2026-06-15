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

interface LocationState {
  from?: string;
}

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as LocationState | null)?.from ?? "/decisions";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<
    Partial<Record<FieldName, string>>
  >({});
  const [formErrors, setFormErrors] = useState<string[]>([]);
  const [pending, setPending] = useState(false);

  const emailRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  function validate(): Partial<Record<FieldName, string>> {
    const errors: Partial<Record<FieldName, string>> = {};
    if (!email.trim()) {
      errors.email = "Email is required.";
    } else if (!EMAIL_PATTERN.test(email.trim())) {
      errors.email = "Please enter a valid email address.";
    }
    if (!password) {
      errors.password = "Password is required.";
    }
    return errors;
  }

  function focusFirstInvalid(errors: Partial<Record<FieldName, string>>) {
    if (errors.email) {
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
    const result = await auth.login(email.trim(), password);
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
        <CardTitle>Welcome back</CardTitle>
        <CardDescription>
          Sign in to your Decision Intelligence OS account.
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

          <FormField label="Password" error={fieldErrors.password} required>
            {(props) => (
              <PasswordInput
                {...props}
                ref={passwordRef}
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            )}
          </FormField>

          <Button type="submit" loading={pending} disabled={pending}>
            Sign in
          </Button>
        </form>

        <p className="mt-4 text-sm text-muted-foreground">
          Need an account?{" "}
          <Link
            to="/register"
            className="font-medium text-primary hover:underline"
          >
            Register
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
