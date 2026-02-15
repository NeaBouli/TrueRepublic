import React, { useState, useEffect } from "react";
import { View, Text, Button, TextInput, StyleSheet, Alert } from "react-native";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

export default function WalletScreen() {
    const [wallet, setWallet] = useState(null);
    const [address, setAddress] = useState(null);
    const [balance, setBalance] = useState("—");
    const [recipient, setRecipient] = useState("");
    const [amount, setAmount] = useState("");
    const [mnemonic, setMnemonic] = useState("");

    const connectWallet = async () => {
        try {
            const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, { prefix: "truerepublic" });
            const accounts = await signer.getAccounts();
            setWallet(signer);
            setAddress(accounts[0].address);
            updateBalance(accounts[0].address);
        } catch (error) {
            Alert.alert("Error", error.message);
        }
    };

    const updateBalance = async (addr) => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const bal = await client.getBalance(addr, "pnyx");
            setBalance(`${bal.amount} PNYX`);
        } catch (error) {
            setBalance("Error fetching balance");
        }
    };

    const sendPNYX = async () => {
        if (!wallet || !recipient || !amount) return Alert.alert("Error", "Please fill all fields.");
        try {
            const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, wallet);
            const result = await client.sendTokens(address, recipient, [{ denom: "pnyx", amount }], "auto");
            Alert.alert("Success", "TX: " + result.transactionHash);
            updateBalance(address);
        } catch (error) {
            Alert.alert("Error", error.message);
        }
    };

    useEffect(() => {
        if (address) {
            const interval = setInterval(() => updateBalance(address), 5000);
            return () => clearInterval(interval);
        }
    }, [address]);

    return (
        <View style={styles.container}>
            <Text style={styles.title}>Wallet</Text>
            {!address ? (
                <>
                    <TextInput
                        style={styles.input}
                        placeholder="Enter mnemonic phrase"
                        placeholderTextColor="#9CA3AF"
                        value={mnemonic}
                        onChangeText={setMnemonic}
                        multiline
                    />
                    <Button title="Connect Wallet" onPress={connectWallet} />
                </>
            ) : (
                <>
                    <Text style={styles.label}>Address: {address}</Text>
                    <Text style={styles.label}>Balance: {balance}</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="Recipient Address"
                        placeholderTextColor="#9CA3AF"
                        value={recipient}
                        onChangeText={setRecipient}
                    />
                    <TextInput
                        style={styles.input}
                        placeholder="Amount (PNYX)"
                        placeholderTextColor="#9CA3AF"
                        value={amount}
                        onChangeText={setAmount}
                        keyboardType="numeric"
                    />
                    <Button title="Send PNYX" onPress={sendPNYX} />
                </>
            )}
        </View>
    );
}

const styles = StyleSheet.create({
    container: { flex: 1, justifyContent: "center", alignItems: "center", backgroundColor: "#1F2937", padding: 20 },
    title: { fontSize: 24, fontWeight: "bold", marginBottom: 20, color: "#fff" },
    label: { fontSize: 14, color: "#D1D5DB", marginBottom: 8 },
    input: { borderWidth: 1, borderColor: "#4B5563", padding: 12, width: "90%", marginBottom: 12, backgroundColor: "#374151", color: "#fff", borderRadius: 8 },
});
