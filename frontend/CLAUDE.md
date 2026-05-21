@AGENTS.md

# CLAUDE.md — Frontend (Expo + React Native)

Scoped conventions for the Expo + React Native app. Inherits the rules in
the root `CLAUDE.md`. Read the root file first.

---

## Project identity reminder

The project is **Vespin**. The app companions
the fictional **Vespin Retro** speaker series. App display name, package
names, and EAS slug all use **`vespin`**.

## Stack — locked

- **Expo** managed workflow
- **Expo Router** for file-based routing (Expo's routing solution, built on
  React Navigation)
- **NativeWind** (`nativewind`) — Tailwind for React Native
- **TanStack Query** (`@tanstack/react-query`) — server state
- **Zustand** — client state
- React **Context** — static globals (theme, i18n)
- **React Hook Form** (`react-hook-form`) + **Zod** (`zod`) — forms
- **Orval** — OpenAPI codegen → TanStack Query hooks
- **expo-localization** — i18n (lightweight wrapper, NOT i18next)
- **expo-secure-store** for the auth token, NOT AsyncStorage

Do NOT add or substitute:
- React Navigation directly (use Expo Router)
- StyleSheet-only styling (use NativeWind)
- Redux, RTK Query, Jotai, Recoil, MobX
- Formik, TanStack Form (use React Hook Form)
- yup, joi, valibot (use Zod)
- Tamagui, styled-components, Restyle, unistyles
- axios as a top-level dependency (Orval generates a fetch-based client by default)
- i18next, react-intl

## Project layout

```
frontend/
├── app/                                Expo Router — ROUTES ONLY
│   ├── _layout.tsx                     Root providers + splash + fonts
│   ├── index.tsx                       Entry: redirects to auth or app
│   ├── (auth)/
│   │   ├── _layout.tsx                 Redirects to (app) if authed
│   │   ├── login.tsx
│   │   ├── register.tsx
│   │   └── guest.tsx
│   ├── (app)/
│   │   ├── _layout.tsx                 Redirects to (auth) if not authed
│   │   ├── (tabs)/
│   │   │   ├── _layout.tsx             Bottom tab bar
│   │   │   ├── devices.tsx
│   │   │   ├── eq.tsx
│   │   │   ├── party.tsx
│   │   │   └── settings.tsx
│   │   ├── devices/[id].tsx
│   │   ├── devices/new.tsx
│   │   ├── eq/[id].tsx
│   │   ├── eq/new.tsx
│   │   ├── party/[id].tsx
│   │   └── party/new.tsx
│   └── +not-found.tsx
├── src/
│   ├── features/                       Domain features — mirror backend
│   │   ├── auth/
│   │   │   ├── components/
│   │   │   ├── hooks/
│   │   │   ├── schemas/                Zod schemas
│   │   │   └── store.ts                Zustand: token, currentUser
│   │   ├── devices/
│   │   ├── eq-profiles/
│   │   ├── party-sessions/
│   │   └── firmware/
│   ├── components/                     Shared, domain-agnostic UI
│   │   ├── ui/                         Button, Input, Card, Modal, Slider
│   │   ├── layout/                     Screen, Section, EmptyState
│   │   └── feedback/                   Toast, Spinner, ErrorMessage
│   ├── api/
│   │   ├── generated/                  Orval output — NEVER EDIT
│   │   ├── client.ts                   Configured fetcher with auth injection
│   │   ├── query-client.ts             TanStack QueryClient setup
│   │   └── error.ts                    API error type + code → message mapper
│   ├── lib/                            Pure utilities, no React
│   ├── hooks/                          App-wide hooks (not feature-specific)
│   ├── providers/                      Context providers
│   ├── theme/                          Design tokens (colors, spacing, typography)
│   ├── i18n/
│   │   ├── translations/
│   │   │   ├── en.json
│   │   │   └── tr.json
│   │   └── index.ts
│   └── types/
├── assets/                             Static images, fonts, icons
├── app.json                            Expo config
├── eas.json                            EAS Build/Update config
├── tailwind.config.js                  NativeWind config — consumes theme tokens
├── tsconfig.json
├── orval.config.ts                     Orval generator config
└── package.json
```

## The `app/` vs `src/` distinction

This is the single most important structural rule:

- **`app/` contains route files only.** Every `.tsx` file in `app/` is a
  route (with rare exceptions like `_layout.tsx` and `+not-found.tsx`).
- **`src/` contains everything else.** Components, hooks, stores, utilities,
  generated API code.

If you need a non-route file in `app/`, you're doing it wrong. Move it to `src/`.

## Screen files are thin

A screen file (`app/(app)/devices/[id].tsx`) does three things and only three:

1. Read route params via `useLocalSearchParams`.
2. Render a feature component, passing the params.
3. Nothing else.

```tsx
// app/(app)/devices/[id].tsx
import { useLocalSearchParams } from "expo-router"
import { DeviceDetail } from "@/features/devices/components/DeviceDetail"

export default function DeviceDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>()
  return <DeviceDetail deviceId={id} />
}
```

NEVER put data fetching in a screen file. NEVER put business logic in a
screen file. NEVER put UI beyond a single component invocation in a screen file.

The reason: routes change, features don't. A feature in `src/features/devices/`
is testable, reusable, and independent of the routing layer.

## Auth gating happens in route group layouts

The `app/(auth)/_layout.tsx` and `app/(app)/_layout.tsx` files handle
redirect-on-auth-state. NEVER add auth checks to individual screen components.

```tsx
// app/(app)/_layout.tsx — schematic
import { Redirect, Stack } from "expo-router"
import { useAuthStore } from "@/features/auth/store"

export default function AppLayout() {
  const token = useAuthStore((s) => s.token)
  if (!token) return <Redirect href="/(auth)/login" />
  return <Stack />
}
```

## Path aliases

The TS path alias `@/*` maps to `src/*`. Configured in `tsconfig.json` AND
`babel.config.js` (via `babel-plugin-module-resolver`) — both must agree.

Use:
```ts
import { DeviceCard } from "@/features/devices/components/DeviceCard"
```

Never:
```ts
import { DeviceCard } from "../../../features/devices/components/DeviceCard"
```

## The API layer

### Generated code is read-only

`src/api/generated/` is produced by **Orval** from `backend/api/openapi.yaml`.

- **NEVER edit files in `src/api/generated/`.** Regenerate instead:
  `pnpm orval`.
- CI enforces this with a drift check. PRs that change the spec without
  regenerating, OR that edit generated files directly, will fail.
- If the generator output is wrong, the fix is in `orval.config.ts` (the
  generator config) or `backend/api/openapi.yaml` (the spec), never in the
  generated files themselves.

### Domain hooks wrap generated hooks

Generated hooks (`useListDevices`, `useCreateDevice`, etc.) are the **raw
material**. Feature code calls **domain hooks** in `src/features/<feature>/hooks/`,
which wrap the generated ones with feature-aware behavior:

```ts
// src/features/devices/hooks/useDevices.ts
import { useListDevices } from "@/api/generated/hooks"

export function useDevices() {
  return useListDevices({
    query: {
      staleTime: 30_000,
      select: (data) => [...data].sort((a, b) =>
        b.createdAt.localeCompare(a.createdAt)
      ),
    },
  })
}
```

Use the domain hook from screens and feature components. Do NOT call
generated hooks directly from screen-level code.

The exception: if a domain hook would be pure passthrough with zero changes,
skip it and call the generated hook directly. Wrap only when there's
something to add (sorting, default `staleTime`, invalidation logic, composition).

### Invalidation on mutations

Mutations should invalidate the affected queries via the TanStack Query
client. This belongs in the domain hook, not in the screen:

```ts
// src/features/eq-profiles/hooks/useForkProfile.ts — schematic
import { useQueryClient } from "@tanstack/react-query"
import { useForkEQProfile } from "@/api/generated/hooks"

export function useForkProfile() {
  const qc = useQueryClient()
  return useForkEQProfile({
    mutation: {
      onSuccess: () => {
        qc.invalidateQueries({ queryKey: ["/eq-profiles"] })
      },
    },
  })
}
```

### The fetch client

`src/api/client.ts` exports the fetcher Orval uses. It:
- Reads the base URL from `EXPO_PUBLIC_API_URL` env var.
- Reads the JWT from the auth Zustand store and injects `Authorization: Bearer ...`.
- Detects 401 responses and triggers a global "log out" action.
- Surfaces the API error envelope `{error: {code, message}}` as a typed error.

NEVER bypass this client. NEVER call `fetch` directly. NEVER call generated
fetch helpers with a hand-written `Authorization` header — the client handles it.

## State management — the three buckets

Be explicit about which kind of state you're dealing with:

| Kind | Tool | Lives in |
|---|---|---|
| Server state (anything from the API) | TanStack Query | Cache, accessed via domain hooks |
| Cross-cutting client state (auth token) | Zustand | `src/features/auth/store.ts` |
| Feature-local client state (modal open, draft form, etc.) | Zustand or `useState` | Feature folder, or component-local |
| Truly static globals (theme, i18n) | React Context | `src/providers/` |

NEVER store server data in Zustand. The cache is TanStack Query. Storing
fetched data in Zustand creates a second source of truth and they will diverge.

NEVER use Context for frequently-changing state. Context causes every
consumer to re-render on every change.

NEVER use `useState` for state that needs to persist across screens. Lift to
a Zustand store in the feature folder.

## Forms

React Hook Form + Zod. Pattern:

```tsx
// src/features/auth/schemas/login.ts
import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
})

export type LoginInput = z.infer<typeof loginSchema>
```

```tsx
// src/features/auth/components/LoginForm.tsx — schematic
const form = useForm<LoginInput>({
  resolver: zodResolver(loginSchema),
  defaultValues: { email: "", password: "" },
})
```

Rules:
- Zod schemas live in `src/features/<feature>/schemas/`.
- The Zod schema is the source of truth for both validation rules AND the
  TS type (`z.infer<typeof schema>`). Do not write a separate TS interface.
- Use `Controller` for inputs from custom UI components that don't accept refs.
- NEVER do validation outside the schema. If the user enters something
  invalid, the schema catches it.

## Styling — NativeWind only

All styling goes through NativeWind className strings:

```tsx
<View className="flex-1 bg-background p-4">
  <Text className="text-lg font-semibold text-foreground">Devices</Text>
</View>
```

Rules:
- NEVER use the StyleSheet API except as a last resort for properties
  NativeWind can't express (rare).
- NEVER use inline `style={...}` objects for static styles. Use className.
- Dynamic styles (e.g., conditional color based on state) use template
  literals: `className={\`p-4 \${isActive ? "bg-primary" : "bg-muted"}\`}`.
- Design tokens (colors, spacing, typography) live in `src/theme/*.ts` and
  are imported by `tailwind.config.js`. NEVER hardcode colors. NEVER use
  raw hex values in className. Always go through the theme.

## Theme and i18n

### Theme

`src/theme/colors.ts`, `src/theme/spacing.ts`, `src/theme/typography.ts`
export the tokens. `tailwind.config.js` imports them and extends the Tailwind
config. Light/dark mode is handled via NativeWind's CSS-variable approach,
toggled by `src/providers/ThemeProvider.tsx`.

When adding a new color: add to `colors.ts`, regenerate (Tailwind picks it
up on next start), use as `bg-<name>` / `text-<name>` etc.

NEVER define colors in components. NEVER use the same color in two places
without going through a token.

### i18n

Strings live in `src/i18n/translations/<lang>.json`. Two languages: `en` and `tr`.

```ts
// Schematic
const { t } = useTranslation()
return <Text>{t("devices.empty")}</Text>
```

Rules:
- ALL user-facing strings go through `t()`. NEVER hardcode English in components.
- Translation keys are dot-namespaced by feature: `devices.empty`,
  `eq.fork_button`, `auth.login.title`.
- When adding a key, add it to BOTH `en.json` and `tr.json`.

## Component conventions

### File naming

- **Components:** PascalCase file, default export. `DeviceCard.tsx` exports `DeviceCard`.
- **Hooks:** camelCase file, named export. `useDevices.ts` exports `useDevices`.
- **Schemas:** camelCase file, named export. `login.ts` exports `loginSchema`.
- **Routes (Expo Router):** lowercase or kebab-case per Expo conventions:
  `devices.tsx`, `[id].tsx`, `_layout.tsx`.

### Component structure

```tsx
type Props = {
  deviceId: string
}

export function DeviceDetail({ deviceId }: Props) {
  const { data, isLoading, error } = useDevice(deviceId)

  if (isLoading) return <Spinner />
  if (error) return <ErrorMessage error={error} />
  if (!data) return <EmptyState />

  return <View>...</View>
}
```

Rules:
- Always type props with a `Props` type alias (not `interface`, just a convention).
- Always handle loading and error states explicitly. Never render against
  undefined `data`.
- Components in `src/components/ui/` MUST be domain-agnostic. If a UI
  component imports from `src/features/`, it's in the wrong folder.

## What goes where

| Code | Location |
|---|---|
| Route definitions (screen files) | `app/` |
| Domain-specific components | `src/features/<feature>/components/` |
| Domain-specific hooks (incl. API hook wrappers) | `src/features/<feature>/hooks/` |
| Domain-specific Zod schemas | `src/features/<feature>/schemas/` |
| Feature-local Zustand store | `src/features/<feature>/store.ts` |
| Domain-agnostic UI primitives | `src/components/ui/` |
| Layout components (Screen, Section) | `src/components/layout/` |
| Toast, Spinner, etc. | `src/components/feedback/` |
| Pure utility functions (no React) | `src/lib/` |
| App-wide hooks (not feature-specific) | `src/hooks/` |
| Context providers | `src/providers/` |
| Design tokens | `src/theme/` |
| API generated code | `src/api/generated/` (read-only) |
| API client + error helpers | `src/api/` |
| Translations | `src/i18n/translations/` |

If you can't decide where something goes, ask before guessing.

## Environment variables

Expo exposes vars prefixed `EXPO_PUBLIC_*` to the JS bundle. The relevant one
is `EXPO_PUBLIC_API_URL`:

- Local dev: `http://<your-lan-ip>:8080` in `frontend/.env.local`.

NEVER read other env vars in the JS bundle. NEVER put secrets in
`EXPO_PUBLIC_*` — they're public.

## Auth token storage

The JWT lives in `expo-secure-store`. The auth Zustand store reads from /
writes to it. NEVER use AsyncStorage for the token — it's not encrypted.

The auth store provides:
- `token`: current JWT or null
- `currentUser`: parsed user from `/users/me` (cached, refreshed periodically)
- `setSession(token, user)`: write both
- `clearSession()`: log out

## Testing

For HCI scope, frontend tests are NOT required. Do not propose:
- Jest setup beyond what `create-expo-app` provides
- React Native Testing Library boilerplate
- Detox or other E2E frameworks
- Snapshot tests
- Coverage targets

If a teammate wants to add tests later, the patterns will be conventional —
just keep components testable by keeping them thin and pure.

## What never to do

- Never edit `src/api/generated/`. Regenerate via `pnpm orval`.
- Never put data fetching, business logic, or UI beyond a single component
  in a screen file under `app/`.
- Never call `fetch` directly. Always go through the configured client.
- Never store server data in Zustand. TanStack Query is the cache.
- Never hardcode colors, spacing, or typography in components. Use theme
  tokens via NativeWind.
- Never hardcode user-facing strings. Use `t()`.
- Never use AsyncStorage for the auth token.
- Never suggest leaving the Expo managed workflow.
- Never bring in a state management library other than the ones listed.
