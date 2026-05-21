# Feature: Design System Foundation

## Summary

Lay the foundation for a Vespin-branded design system in the Expo + React Native
frontend: a two-tier token sheet (primitive → semantic), an `AppText` primitive
backed by ZenDots (display) + platform-default system font (body), a small set
of layout/UI primitives (`Screen`, `Button`, `Mark`, `Icon`), a Surface tone
context for inverted (maroon) surfaces, custom font loading via `expo-font`,
asset reorganization, and a custom bottom tab bar matching the new designs.

Scope is **foundation only** — token sheet, primitives, conventions, and the
minimum routing changes to host the new auth flow. Feature-level screens
(Welcome content, Login form restyle, Home empty/selected, custom tab bar
visuals) are explicitly downstream of this spec.

Constraints honored: Expo managed workflow, no native builds, no new state
libraries, NativeWind for all styling, no dark mode (designs don't define one),
no Storybook/visual regression (HCI scope).

## Requirements

1. Add ZenDots as the brand display face via `expo-font`, loaded once at
   app boot; OS splash held until ready via `expo-splash-screen`.
2. Body text uses each platform's system font (no hardcoded family).
3. Provide a two-tier color token system: primitive palette (`brand.*`,
   `neutral.*`) → semantic aliases (`background`, `surface`, `primary`,
   `onPrimary`, `border`, `muted`, `danger`).
4. Token sheet drives `tailwind.config.js`; raw hex/size values are forbidden
   in component className strings.
5. Ship an `AppText` primitive with a minimal variant set (`display`,
   `button`, `title`, `body`, `caption`) and a `tone` prop (`default`,
   `muted`, `onPrimary`, `danger`). Display + button variants use ZenDots;
   the rest use system. Author controls casing — no auto-transform.
6. Ship a `Screen` layout primitive with a `tone` prop (`default` = cream,
   `primary` = maroon) that sets background AND publishes a Surface context
   so nested `AppText` flips its default tone automatically.
7. Ship a `Button` primitive with `variant` (`primary`, `secondary`, `ghost`)
   and `size` (`md`, `lg`) props; labels render through `AppText`.
8. Ship a `Mark` brand component (Vespin logo at `sm`/`md`/`lg`) and an
   `Icon` wrapper around `@expo/vector-icons` (Ionicons).
9. Enforce a 44pt minimum hit area on `Button` and `IconButton`. Icon-only
   controls require an `accessibilityLabel` prop (TypeScript-enforced).
10. Variants are expressed as plain object maps in each primitive's file.
    No `tailwind-variants`, `cva`, or similar.
11. Components in `src/components/ui/` and `src/components/layout/` MUST NOT
    import from `src/features/`. Codified in `frontend/CLAUDE.md`.
12. Move `ZenDots-Regular.ttf` → `frontend/assets/fonts/`. Move `hero-1.png`,
    `hero-2.png` → `frontend/assets/images/`. Brand mark assets live in
    `frontend/assets/brand/`.
13. Add a `(auth)/index.tsx` route as the Welcome screen and a
    `(auth)/chooser.tsx` route for the Login / Sign up selection. `index.tsx`
    becomes the redirect entry for the auth group.
14. Replace the default `Tabs` tab bar with a custom one consumed via
    `tabBar={(props) => <BottomTabBar {...props} />}`. Tab bar is icon-only,
    3 tabs (home, eq, settings). The Party tab is removed from the bar;
    Party remains reachable via stack routes from home/device detail.
15. No new top-level dependencies beyond `expo-font`, `expo-splash-screen`,
    and `@expo/vector-icons` (the last two ship with Expo).

## Data Model Changes

None. This is frontend-only foundation work.

## API Endpoints

None.

## Module Structure

### Files to create

```
frontend/
├── assets/
│   ├── fonts/
│   │   └── ZenDots-Regular.ttf            (moved from repo root)
│   ├── images/
│   │   ├── hero-welcome.png               (renamed from hero-1.png)
│   │   └── hero-auth.png                  (renamed from hero-2.png)
│   └── brand/
│       └── mark.png                       (placeholder; final asset TBD)
├── src/
│   ├── theme/
│   │   ├── colors.ts                      Primitive + semantic color tokens
│   │   ├── spacing.ts                     Spacing scale
│   │   ├── radius.ts                      Radii (incl. `pill`)
│   │   ├── typography.ts                  Font family map + type scale
│   │   └── index.ts                       Barrel
│   ├── components/
│   │   ├── ui/
│   │   │   ├── AppText.tsx
│   │   │   ├── Button.tsx
│   │   │   ├── IconButton.tsx
│   │   │   └── Icon.tsx                   Ionicons wrapper
│   │   ├── layout/
│   │   │   ├── Screen.tsx                 Surface tone + safe area + bg
│   │   │   ├── Section.tsx
│   │   │   └── EmptyState.tsx
│   │   ├── brand/
│   │   │   └── Mark.tsx
│   │   └── feedback/                      (existing folder, untouched here)
│   └── providers/
│       └── SurfaceProvider.tsx            Context for current surface tone
└── app/
    └── (auth)/
        ├── index.tsx                      NEW: Welcome route (slot for hero)
        └── chooser.tsx                    NEW: Login / Sign-up chooser
```

### Files to modify

```
frontend/
├── app/_layout.tsx                        Load ZenDots, hold splash until ready
├── app/index.tsx                          Redirect target updates (auth → /(auth))
├── app/(auth)/_layout.tsx                 No behavioral change; default Stack
│                                          options pick up Screen tone via children
├── app/(app)/(tabs)/_layout.tsx           Use custom BottomTabBar; drop party tab
├── tailwind.config.js                     Consume tokens from src/theme/
├── global.css                             Token-driven CSS vars (NativeWind)
├── app.json                               Splash config (maroon bg + mark)
└── tsconfig.json                          (no change expected; @/* already wired)
```

### Files to delete

```
app/(app)/(tabs)/party.tsx                 Tab removed; route lives at
                                           app/(app)/party/* already
```

### Documentation updates

Add a short section to `frontend/CLAUDE.md` titled **"Design system rules"**
covering:

- Two-tier token system; only primitives in `theme/`, only semantics in
  components.
- Components in `ui/` and `layout/` cannot import from `features/`.
- All text goes through `AppText`. No bare `<Text>` in features or screens.
- All colors via semantic Tailwind classes (`bg-surface`, `text-onPrimary`).
  No raw hex in className.
- Display + button variants use ZenDots; all other text uses system.
- Icons go through `<Icon name="..." />`, which wraps Ionicons.

## Business Logic

### Token sheet (provisional values; refine when official hexes arrive)

`src/theme/colors.ts`:

```ts
export const primitive = {
  brand: {
    maroon: { 50: "#F7E9EC", 500: "#7A1424", 700: "#5C0F1A", 900: "#3A0911" },
    cream:  { 50: "#F5EFE3", 100: "#EFE9DE", 200: "#E4DCCB" },
  },
  neutral: { 0: "#FFFFFF", 900: "#0A0A0A", 700: "#3A3A3A", 500: "#6B6B6B", 300: "#B8B0A2" },
  status:  { danger: "#B3261E", success: "#2E7D32" },
} as const;

export const semantic = {
  background:   primitive.brand.cream[100],
  surface:      primitive.brand.cream[50],
  surfaceAlt:   primitive.brand.cream[200],
  primary:      primitive.brand.maroon[700],
  primaryDeep:  primitive.brand.maroon[900],
  onPrimary:    primitive.brand.cream[50],
  ink:          primitive.neutral[900],
  muted:        primitive.neutral[500],
  border:       primitive.neutral[300],
  danger:       primitive.status.danger,
} as const;
```

`src/theme/typography.ts`:

```ts
export const fontFamily = {
  display: "ZenDots",      // Loaded via expo-font; ZenDots-Regular.ttf
  body:    undefined,      // Platform default (SF on iOS, Roboto on Android)
} as const;

export const typeScale = {
  display: { size: 40, lineHeight: 44, weight: "400", family: "display" },
  button:  { size: 16, lineHeight: 20, weight: "400", family: "display" },
  title:   { size: 20, lineHeight: 28, weight: "600", family: "body" },
  body:    { size: 16, lineHeight: 24, weight: "400", family: "body" },
  caption: { size: 13, lineHeight: 18, weight: "400", family: "body" },
} as const;
```

`tailwind.config.js` imports these and exposes them as utilities:
`bg-background`, `bg-surface`, `bg-primary`, `text-onPrimary`, `text-ink`,
`text-muted`, `border-border`, `font-display`, `font-body`, `rounded-pill`.

### Font loading flow

In `app/_layout.tsx`:

1. Call `SplashScreen.preventAutoHideAsync()` at module top.
2. `useFonts({ ZenDots: require("@/assets/fonts/ZenDots-Regular.ttf") })`.
3. While `!loaded && !error`: return `null` (splash stays visible).
4. On `loaded || error`: `SplashScreen.hideAsync()` in a `useEffect`, then
   render the provider tree.
5. On `error`: render the tree anyway. ZenDots falls back to system font;
   the app remains usable. Log via `console.warn`.

### `AppText` variant resolution

```ts
const VARIANT = {
  display: "font-display text-[40px] leading-[44px]",
  button:  "font-display text-base leading-5",
  title:   "font-body text-xl leading-7 font-semibold",
  body:    "font-body text-base leading-6",
  caption: "font-body text-[13px] leading-[18px]",
} as const;

const TONE = {
  default:   "text-ink",
  muted:     "text-muted",
  onPrimary: "text-onPrimary",
  danger:    "text-danger",
} as const;
```

Resolution order at render:

1. Read `tone` prop. If undefined, read from `SurfaceContext` (default for
   `Screen tone="primary"` is `"onPrimary"`; for `tone="default"` it's
   `"default"`).
2. Resolve `variant` → className string.
3. Compose: `${VARIANT[variant]} ${TONE[resolvedTone]} ${className ?? ""}`.

The `className` prop is appended last so callers can layer one-off spacing
or weight overrides without forking a variant.

### Surface tone context

```ts
type SurfaceTone = "default" | "primary";
const SurfaceContext = createContext<SurfaceTone>("default");
```

`Screen` reads its `tone` prop and wraps children in
`<SurfaceContext.Provider value={tone}>`. Default tone is `"default"` so any
`AppText` rendered outside a `Screen` (tests, error boundaries) still resolves
to a sensible color. No throw, no required provider.

`AppText` consumes the context only as a fallback — explicit `tone` prop
always wins. This lets a cream chip live on a maroon screen without
prop-drilling.

### `Button` variants

```ts
const BTN_VARIANT = {
  primary:   "bg-primary",                    // label tone: onPrimary
  secondary: "bg-surface border border-border", // label tone: ink
  ghost:     "bg-transparent",                // label tone: ink
} as const;

const BTN_SIZE = {
  md: "h-11 px-4 rounded-pill",   // 44pt min hit
  lg: "h-14 px-6 rounded-pill",
} as const;
```

Label is always rendered as `<AppText variant="button" tone={...} />`. Label
tone is derived from `variant` (primary → onPrimary; secondary/ghost → ink)
unless explicitly overridden via a `labelTone` prop.

### Icon system

`Icon` is a thin wrapper around `Ionicons` that takes a known name set
(narrowed `IconName` union) and a `size`/`color` derived from theme. Apps
use `<Icon name="settings-outline" size="md" tone="onPrimary" />` rather than
reaching for raw Ionicons. This gives one chokepoint to swap families later.

`IconButton` composes `Icon` with a 44pt pressable wrapper and a required
`accessibilityLabel` prop. TypeScript enforces the label via a required key
in `Props` — no runtime check needed.

### Custom bottom tab bar

`src/components/layout/BottomTabBar.tsx` is a function component conforming
to `BottomTabBarProps` from React Navigation. It:

- Renders a maroon bar (`bg-primary`) with safe-area bottom inset.
- Maps each route to an icon (route name → icon name table at top of file).
- Active state via tone (`onPrimary` vs `onPrimary/60`); no labels.
- Calls `navigation.navigate(route.name)` on press.

Wired in `(tabs)/_layout.tsx`:

```tsx
<Tabs tabBar={(props) => <BottomTabBar {...props} />}
      screenOptions={{ headerShown: false }}>
  <Tabs.Screen name="devices"  options={{ title: "Home" }} />
  <Tabs.Screen name="eq"       options={{ title: "EQ" }} />
  <Tabs.Screen name="settings" options={{ title: "Settings" }} />
</Tabs>
```

(Tab `name` matches the existing file `devices.tsx`. Renaming the file to
`index.tsx` is deferred — out of scope for the foundation.)

### Splash → Welcome handoff

- `app.json` `expo.splash` configured with maroon background color and the
  brand mark PNG centered.
- OS shows splash until JS bundle ready AND `expo-font` resolves AND
  `SplashScreen.hideAsync()` is called.
- `app/index.tsx` redirects: not hydrated → `null` (splash still up);
  no token → `/(auth)` (which resolves to `(auth)/index.tsx` = Welcome);
  token present → `/(app)/(tabs)/devices`.

## Edge Cases & Error Handling

| Case | Behavior |
|---|---|
| ZenDots font fails to load | Hide splash anyway, render app, log warn. Display variants fall back to system; layout doesn't break because line-height is fixed. |
| `AppText` rendered with no `Screen` ancestor | Surface context default is `"default"`; text renders `text-ink` on whatever background is mounted. No throw. |
| `Screen tone="primary"` nested inside another `Screen tone="default"` | Inner provider value wins for its subtree. AppText inside the nested Screen flips to `onPrimary` automatically. Nesting works correctly without effort. |
| Caller passes raw hex via `className="bg-#5C0F1A"` | Tailwind will not recognize it (no JIT for arbitrary hex without `bg-[#...]` syntax); convention is enforced by code review, not the type system. Document in CLAUDE.md. |
| Caller passes `className` to `AppText` that conflicts with variant (e.g., `text-xs` on `variant="display"`) | Caller's className wins because it's appended last. This is intentional — variants are defaults, not locks. |
| `IconButton` missing `accessibilityLabel` | TypeScript compile error (required prop). |
| OS font scaling (Dynamic Type) very large | `AppText` does not set `maxFontSizeMultiplier`. Layout components must use flexible heights, not fixed; `Button` `h-11`/`h-14` traded for `min-h-*` in a follow-up if a screen actually breaks. |
| `expo-splash-screen` `hideAsync` called twice | No-op on second call; safe to call from a `useEffect` that may re-run. |
| Hero PNGs missing | Image components render with `defaultSource` empty; surface stays maroon. Documented as expected during foundation phase. |

## Security Considerations

Not applicable. Design system contains no data flow, no auth surface, no
network calls. The only "input" is the theme tokens themselves, which are
compile-time constants.

The Surface context is purely visual and cannot be used to leak data; it
holds a string literal type.

## Testing Plan

Per project posture (HCI scope, frontend tests not required), no automated
tests are required for the foundation. If the team chooses to add a few:

**Given/when/then sketches** for the only non-trivial logic:

1. **AppText tone resolution**
   - *Given* an `AppText variant="body"` with no `tone` prop
   - *When* rendered inside `<Screen tone="primary">`
   - *Then* the resolved className includes `text-onPrimary`

2. **AppText explicit tone wins over context**
   - *Given* an `AppText variant="body" tone="ink"`
   - *When* rendered inside `<Screen tone="primary">`
   - *Then* the resolved className includes `text-ink`, not `text-onPrimary`

3. **Font load failure does not block app**
   - *Given* `useFonts` returns `[false, Error]`
   - *When* root layout renders
   - *Then* `SplashScreen.hideAsync` is still called and children mount

The first two are pure-function tests on the variant resolver — extract
`resolveTextClasses(variant, tone, contextTone, className?)` from `AppText`
into a sibling module to make this trivial. Worth doing for clarity alone.

Manual verification checklist for the foundation PR:

- [ ] App launches; splash is maroon with mark; transitions to Welcome route.
- [ ] ZenDots renders on display text (visible weight/shape difference vs system).
- [ ] iOS shows SF Pro for body; Android shows Roboto for body.
- [ ] Auth screens have maroon background; tab screens have cream background.
- [ ] Bottom tab bar renders 3 icons on cream surface, maroon bar.
- [ ] No raw hex strings in any `className` (grep `className=".*#[0-9A-F]`).
- [ ] No `<Text>` imports in `src/features/` or `app/` (grep `from "react-native".*Text`).

## Open Questions

1. **Final brand hex values** — provisional palette is sampled from the
   PNGs. User to supply authoritative `brand.maroon` and `brand.cream` ramps.
   When swapping in, only `src/theme/colors.ts` changes; no component touch
   needed.
2. **Brand mark asset** — current `assets/brand/mark.png` is a placeholder
   slot. Need final mark in 1x/2x/3x or as SVG (would require
   `react-native-svg`, which is fine in managed workflow but not assumed in
   this spec).
3. **Display font scale beyond 40pt** — designs show one display size
   (~40pt) but the "WELCOME" overlay reads larger relative to viewport. May
   need a `displayLg` (~56pt) once the Welcome screen is laid out for real.
   Defer until a screen needs it.
4. **Home tab route name** — currently `devices.tsx`. Tab is now called
   "Home" in designs. Renaming the file would let us avoid a label/route
   mismatch but breaks existing imports. Foundation keeps the filename;
   rename is a follow-up.
5. **Party access path** — confirmed Party is no longer a tab. The
   triggering UI (button on home, in device detail, or in settings) is a
   feature-level decision, not a design-system one.
6. **OAuth social row on chooser** — out of scope for the design system,
   but the chooser route will need stub buttons (per root CLAUDE.md, UI
   only). Handled when the chooser screen is built.
