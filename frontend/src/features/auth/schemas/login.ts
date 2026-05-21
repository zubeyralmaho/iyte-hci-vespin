import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().trim().pipe(z.email("auth.errors.invalidEmail")),
  password: z.string().min(1, "auth.errors.passwordRequired"),
});

export type LoginInput = z.infer<typeof loginSchema>;
