import { Pressable, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Icon, type IconName } from "@/components/ui/Icon";

type TabRoute = { key: string; name: string };
type TabNavigation = {
  emit: (event: {
    type: "tabPress";
    target: string;
    canPreventDefault: true;
  }) => { defaultPrevented: boolean };
  navigate: (name: string) => void;
};
type BottomTabBarProps = {
  state: { index: number; routes: TabRoute[] };
  navigation: TabNavigation;
};
import { SurfaceProvider } from "@/providers/SurfaceProvider";

const ROUTE_ICON: Record<string, { active: IconName; inactive: IconName; label: string }> = {
  devices: { active: "home", inactive: "home-outline", label: "Home" },
  eq: { active: "options", inactive: "options-outline", label: "EQ" },
  settings: { active: "settings", inactive: "settings-outline", label: "Settings" },
};

export function BottomTabBar({ state, navigation }: BottomTabBarProps) {
  const insets = useSafeAreaInsets();

  return (
    <SurfaceProvider tone="primary">
      <View
        className="flex-row items-center justify-around bg-primary px-4"
        style={{ paddingBottom: insets.bottom + 8, paddingTop: 12 }}
      >
        {state.routes.map((route, index) => {
          const meta = ROUTE_ICON[route.name];
          if (!meta) return null;
          const isFocused = state.index === index;

          return (
            <Pressable
              key={route.key}
              accessibilityRole="button"
              accessibilityState={isFocused ? { selected: true } : {}}
              accessibilityLabel={meta.label}
              onPress={() => {
                const event = navigation.emit({
                  type: "tabPress",
                  target: route.key,
                  canPreventDefault: true,
                });
                if (!isFocused && !event.defaultPrevented) {
                  navigation.navigate(route.name);
                }
              }}
              className="h-11 w-11 items-center justify-center rounded-pill"
              style={{ opacity: isFocused ? 1 : 0.6 }}
            >
              <Icon
                name={isFocused ? meta.active : meta.inactive}
                size="lg"
                tone="onPrimary"
              />
            </Pressable>
          );
        })}
      </View>
    </SurfaceProvider>
  );
}
