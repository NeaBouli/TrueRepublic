import React, { useState, useEffect } from "react";
import { View, Text, FlatList, StyleSheet, TouchableOpacity } from "react-native";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

export default function GovernanceScreen() {
    const [domains, setDomains] = useState([]);
    const [selectedDomain, setSelectedDomain] = useState(null);

    const fetchDomains = async () => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const result = await client.queryAbci("custom/truedemocracy/domains", new Uint8Array());
            const decoded = new TextDecoder().decode(result.value);
            setDomains(JSON.parse(decoded));
        } catch (err) {
            console.error("Failed to fetch domains:", err);
        }
    };

    useEffect(() => {
        fetchDomains();
    }, []);

    const renderDomain = ({ item }) => (
        <TouchableOpacity
            style={[styles.card, selectedDomain === item.name && styles.cardSelected]}
            onPress={() => setSelectedDomain(selectedDomain === item.name ? null : item.name)}
        >
            <Text style={styles.cardTitle}>{item.name}</Text>
            <Text style={styles.cardSub}>Members: {(item.members || []).length}</Text>
            <Text style={styles.cardSub}>Issues: {(item.issues || []).length}</Text>
            {selectedDomain === item.name && (item.issues || []).map((issue, i) => (
                <View key={i} style={styles.issue}>
                    <Text style={styles.issueName}>{issue.name} — {issue.stones} stones</Text>
                    {(issue.suggestions || []).map((s, j) => (
                        <Text key={j} style={styles.suggestion}>  {s.name} ({s.color}) — {s.stones} stones</Text>
                    ))}
                </View>
            ))}
        </TouchableOpacity>
    );

    return (
        <View style={styles.container}>
            <Text style={styles.title}>Governance</Text>
            <FlatList
                data={domains}
                keyExtractor={(item) => item.name}
                renderItem={renderDomain}
                contentContainerStyle={styles.list}
            />
        </View>
    );
}

const styles = StyleSheet.create({
    container: { flex: 1, backgroundColor: "#1F2937", padding: 16 },
    title: { fontSize: 24, fontWeight: "bold", color: "#fff", textAlign: "center", marginBottom: 16 },
    list: { paddingBottom: 20 },
    card: { backgroundColor: "#374151", padding: 16, borderRadius: 8, marginBottom: 12 },
    cardSelected: { borderColor: "#3B82F6", borderWidth: 2 },
    cardTitle: { fontSize: 18, fontWeight: "bold", color: "#fff" },
    cardSub: { fontSize: 14, color: "#9CA3AF", marginTop: 4 },
    issue: { marginTop: 8, paddingLeft: 12 },
    issueName: { fontSize: 14, color: "#D1D5DB" },
    suggestion: { fontSize: 12, color: "#9CA3AF" },
});
