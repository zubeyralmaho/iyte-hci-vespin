import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiError } from "@/api/error";
import { postAuthGuest, postAuthLogin, postAuthRegister } from "@/api/generated/auth/auth";
import type { AuthResponse } from "@/api/generated/schemas";
import { useAuthStore } from "@/features/auth/store";
import type { LoginInput } from "../schemas/login";
import type { RegisterInput } from "../schemas/register";

function assertAuthResponse(response: { status: number; data: unknown }): AuthResponse {
  if (response.status < 200 || response.status >= 300 || !isAuthResponse(response.data)) {
    throw new ApiError(
      response.status,
      "unexpected_response",
      "Auth API did not return a valid session. Check EXPO_PUBLIC_API_URL.",
    );
  }
  return response.data as AuthResponse;
}

function isAuthResponse(value: unknown): value is AuthResponse {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<AuthResponse>;
  return (
    typeof candidate.token === "string" &&
    !!candidate.user &&
    typeof candidate.user === "object" &&
    typeof candidate.user.id === "string"
  );
}

export function useLogin() {
  const queryClient = useQueryClient();
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async (input: LoginInput) => {
      const auth = assertAuthResponse(await postAuthLogin(input));
      await setSession(auth.token, auth.user);
      queryClient.clear();
      return auth.user;
    },
  });
}

export function useRegister() {
  const queryClient = useQueryClient();
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async (input: RegisterInput) => {
      const displayName = input.displayName?.trim();
      const auth = assertAuthResponse(
        await postAuthRegister({
          email: input.email,
          password: input.password,
          ...(displayName ? { displayName } : {}),
        }),
      );
      await setSession(auth.token, auth.user);
      queryClient.clear();
      return auth.user;
    },
  });
}

export function useContinueAsGuest() {
  const queryClient = useQueryClient();
  const setSession = useAuthStore((s) => s.setSession);

  return useMutation({
    mutationFn: async () => {
      const auth = assertAuthResponse(await postAuthGuest());
      await setSession(auth.token, auth.user);
      queryClient.clear();
      return auth.user;
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  const clearSession = useAuthStore((s) => s.clearSession);

  return useMutation({
    mutationFn: async () => {
      await clearSession();
      queryClient.clear();
    },
  });
}
