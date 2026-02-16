import React from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import WalletScreen from "./screens/WalletScreen";
import GovernanceScreen from "./screens/GovernanceScreen";
import DexScreen from "./screens/DexScreen";

const Tab = createBottomTabNavigator();

export default function App() {
    return (
        <NavigationContainer>
            <Tab.Navigator
                screenOptions={{
                    headerStyle: { backgroundColor: "#1F2937" },
                    headerTintColor: "#fff",
                    headerTitle: "TrueRepublic",
                    tabBarStyle: { backgroundColor: "#1F2937", borderTopColor: "#374151" },
                    tabBarActiveTintColor: "#3B82F6",
                    tabBarInactiveTintColor: "#9CA3AF",
                }}
            >
                <Tab.Screen name="Wallet" component={WalletScreen} />
                <Tab.Screen name="Governance" component={GovernanceScreen} />
                <Tab.Screen name="DEX" component={DexScreen} />
            </Tab.Navigator>
        </NavigationContainer>
    );
}
