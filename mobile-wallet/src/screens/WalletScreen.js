import React, { useState, useEffect } from "react";
import { View, Text, Button, TextInput, StyleSheet } from "react-native";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

export default function WalletScreen() {
    const [wallet, setWallet] = useState(null);
    const [balance, setBalance] = useState("Loading...");
    const [recipient, setRecipient] = useState("");
    const [amount, setAmount] = useState("");

    const connectWallet = async () => {
        try {
            await window.keplr.enable("truerepublic-1");
            const offlineSigner = window.keplr.getOfflineSigner("truerepublic-1");
            const accounts = await offlineSigner.getAccounts();
            setWallet(accounts[0].address);
            updateBalance(accounts[0].address);
        } catch (error) {
            setBalance("Error: " + error.message);
        }
    };

    const updateBalance = async (address) => {
        const client = await SigningStargateClient.connect(RPC_ENDPOINT);
        const balance = await client.getBalance(address, "pnyx");
        setBalance(`${balance.amount} PNYX`);
    };

    const sendPNYX = async () => {
        if (!wallet || !recipient || !amount) return alert("Please fill all fields.");
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, window.keplr.getOfflineSigner("truerepublic-1"));
        const result = await client.sendTokens(wallet, recipient, [{ denom: "pnyx", amount }], "auto");
        alert("Transaction successful: " + result.transactionHash);
        updateBalance(wallet);
    };

    useEffect(() => {
        if (wallet) {
            const interval = setInterval(() => updateBalance(wallet), 5000);
            return () => clearInterval(interval);
        }
    }, [wallet]);

    return (
        <View style={styles.container}>
            <Text style={styles.title}>Wallet</Text>
            <Button title="Connect Wallet" onPress={connectWallet} />
            {wallet && (
                <>
                    <Text>Address: {wallet}</Text>
                    <Text>Balance: {balance}</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="Recipient Address"
                        value={recipient}
                        onChangeText={setRecipient}
                    />
                    <TextInput
                        style={styles.input}
                        placeholder="Amount (PNYX)"
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
    container: { flex: 1, justifyContent: "center", alignItems: "center", backgroundColor: "#1F2937" },
    title: { fontSize: 20, fontWeight: "bold", marginBottom: 10, color: "#fff" },
    input: { borderWidth: 1, padding: 10, width: "80%", marginBottom: 10, backgroundColor: "#374151", color: "#fff" },
});
