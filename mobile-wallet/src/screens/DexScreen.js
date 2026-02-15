import React, { useState, useEffect } from "react";
import { View, Text, TextInput, Button, StyleSheet, Alert } from "react-native";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

export default function DexScreen() {
    const [pools, setPools] = useState([]);
    const [fromAsset, setFromAsset] = useState("pnyx");
    const [toAsset, setToAsset] = useState("atom");
    const [amount, setAmount] = useState("");

    const fetchPools = async () => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const result = await client.queryAbci("custom/dex/pools", new Uint8Array());
            const decoded = new TextDecoder().decode(result.value);
            setPools(JSON.parse(decoded));
        } catch (err) {
            console.error("Failed to fetch pools:", err);
        }
    };

    useEffect(() => {
        fetchPools();
    }, []);

    const swapTokens = async () => {
        if (!amount) return Alert.alert("Error", "Please enter an amount.");
        if (fromAsset === toAsset) return Alert.alert("Error", "From and To must differ.");
        Alert.alert("Info", "Swap requires a connected wallet. Use the Wallet tab to import your mnemonic first.");
    };

    return (
        <View style={styles.container}>
            <Text style={styles.title}>DEX</Text>

            {pools.length > 0 && (
                <View style={styles.section}>
                    <Text style={styles.sectionTitle}>Pools</Text>
                    {pools.map((p, i) => (
                        <Text key={i} style={styles.poolText}>
                            PNYX/{p.asset_denom}: {p.pnyx_reserve} / {p.asset_reserve}
                        </Text>
                    ))}
                </View>
            )}

            <View style={styles.section}>
                <Text style={styles.sectionTitle}>Swap</Text>
                <Text style={styles.label}>From: {fromAsset.toUpperCase()}</Text>
                <Button title={fromAsset === "pnyx" ? "Switch to ATOM" : "Switch to PNYX"} onPress={() => {
                    setFromAsset(fromAsset === "pnyx" ? "atom" : "pnyx");
                    setToAsset(fromAsset === "pnyx" ? "pnyx" : "atom");
                }} />
                <Text style={styles.label}>To: {toAsset.toUpperCase()}</Text>
                <TextInput
                    style={styles.input}
                    placeholder="Amount"
                    placeholderTextColor="#9CA3AF"
                    value={amount}
                    onChangeText={setAmount}
                    keyboardType="numeric"
                />
                <Text style={styles.fee}>Fee: 0.3% swap fee. 1% burn on PNYX output.</Text>
                <Button title="Swap" onPress={swapTokens} />
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    container: { flex: 1, backgroundColor: "#1F2937", padding: 16 },
    title: { fontSize: 24, fontWeight: "bold", color: "#fff", textAlign: "center", marginBottom: 16 },
    section: { backgroundColor: "#374151", padding: 16, borderRadius: 8, marginBottom: 16 },
    sectionTitle: { fontSize: 18, fontWeight: "bold", color: "#fff", marginBottom: 8 },
    label: { fontSize: 14, color: "#D1D5DB", marginVertical: 4 },
    poolText: { fontSize: 13, color: "#9CA3AF", marginBottom: 4 },
    input: { borderWidth: 1, borderColor: "#4B5563", padding: 12, backgroundColor: "#1F2937", color: "#fff", borderRadius: 8, marginVertical: 8 },
    fee: { fontSize: 11, color: "#6B7280", marginBottom: 8 },
});
