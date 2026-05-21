import { z } from "zod";

export const registerSchema = z.object({
  email: z.string().trim().pipe(z.email("auth.errors.invalidEmail")),
  password: z.string().min(8, "auth.errors.passwordLength"),
  displayName: z.string().trim().optional(),
});

export type RegisterInput = z.infer<typeof registerSchema>;
