using System;
using System.Collections.Generic;
using System.Text;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using Pnyx.ApiClient.Solana.Local;
using Pnyx.ApiClient.Solana.User;
using Solnet.KeyStore;
using Solnet.Programs;
using Solnet.Rpc;
using Solnet.Rpc.Builders;
using Solnet.Wallet;
using Solnet.Wallet.Bip39;
using Pnyx.ApiClient.Solana.UnitTests.CodeSamples.MultisigExamples;

namespace Pnyx.ApiClient.Solana.UnitTests
{
    [TestClass]
    public class Local
    {
        private string _securePassword = @"GDKE238gjt={xQh_/Ctfm/2g5/QshdAk";

        [TestMethod]
        public void WriteAndReadFromKeyStore()
        {
            PnyxAccount account = new PnyxAccount("Ingbert");
            string json = KeyStore.Instance.Encrypt(account, _securePassword);

            PnyxAccount decryptedAccount = KeyStore.Instance.Decrypt(json, _securePassword);

            Assert.AreEqual(account.UserName, decryptedAccount.UserName);
        }

        /// <summary>
        /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/SolletKeygenKeystore.cs
        /// </summary>
        [TestMethod]
        public void SolnetKeystoreExample()
        {
            var expectedSolletAddresses = new List<string[]>
            {
                new []{"6bhhceZToGG9RsTe1nfNFXEMjavhj6CV55EsvearAt2z", "5S1UT7L6bQ8sVaPjpJyYFEEYh8HAXRXPFUEuj6kHQXs6ZE9F6a2wWrjdokAmSPP5HVP46bYxsrU8yr2FxxYmVBi6"},
                new []{"9we6kjtbcZ2vy3GSLLsZTEhbAqXPTRvEyoxa8wxSqKp5", "22J7rH3DFJb1yz8JuWUWfrpsQrNsvZKov8sznfwHbPGTznSgQ8u6LQ6KixPC2mYCJDsfzME1FbdX1x89zKq4MU3K"},
                new []{"3F2RNf2f2kWYgJ2XsqcjzVeh3rsEQnwf6cawtBiJGyKV", "5954a6aMxVnPTyMNdVKrSiqoVMRvZcwU7swGp9kHsV9HP9Eu81TebS4Mbq5ZGmZwUaJkkKoCJ2eJSY9cTdWzRXeF"},
                new []{"GyWQGKYpvzFmjhnG5Pfw9jfvgDM7LB31HnTRPopCCS9", "tUV1EeY6CARAbuEfVqKS46X136PRBea8PcmYfHRWNQc6yYB14GkSBZ6PTybUt5W14A7FSJ6Mm6NN22fLhUhDUGu"},
                new []{"GjtWSPacUQFVShQKRKPc342MLCdNiusn3WTJQKxuDfXi", "iLtErFEn6w5xbsUW63QLYMTJeX8TAgFTUDTgat3gxpaiN3AJbebv6ybtmTj1t1yvkqqY2k1uwFxaKZoCQAPcDZe"},
                new []{"DjGCyxjGxpvEo921Ad4tUUWquiRG6dziJUCk8HKZoaKK", "3uvEiJiMyXqQmELLjxV8r3E7CyRFg42LUAxzz6q7fPhzTCxCzPkaMCQ9ARpWYDNiDXhue2Uma1C7KR9AkiiWUS8y"},
                new []{"HU6aKFapq4RssJqV96rfE7vv1pepz5A5miPAMxGFso4X", "4xFZDEhhw3oVewE3UCvzLmhRWjjcqvVMxuYiETWiyaV2wJwEJ4ceDDE359NMirh43VYisViHAwsXjZ3F9fk6dAxB"},
                new []{"HunD57AAvhBiX2SxmEDMbrgQ9pcqrtRyWKy7dWPEWYkJ", "2Z5CFuVDPQXxrB3iw5g6SAnKqApE1djAqtTZDA83rLZ1NDi6z13rwDX17qdyUDCxK9nDwKAHdVuy3h6jeXspcYxA"},
                new []{"9KmfMX4Ne5ocb8C7PwjmJTWTpQTQcPhkeD2zY35mawhq", "c1BzdtL4RByNQnzcaUq3WuNLuyY4tQogGT7JWwy4YGBE8FGSgWUH8eNJFyJgXNYtwTKq4emhC4V132QX9REwujm"},
                new []{"7MrtfwpJBw2hn4eopB2CVEKR1kePJV5kKmKX3wUAFsJ9", "4skUmBVmaLoriN9Ge8xcF4xQFJmF554rnRRa2u1yDbre2zj2wUpgCXUaPETLSAWNudCkNAkWM5oJFJRaeZY1g9JR"}
            };
            const string mnemonicWords = "route clerk disease box emerge airport loud waste attitude film army tray forward deal onion eight catalog surface unit card window walnut wealth medal";

            Mnemonic mnemonic = new Mnemonic(mnemonicWords, WordList.English);

            SecretKeyStoreService keystoreService = new SecretKeyStoreService();

            // no passphrase to generate same keys as sollet
            Wallet wallet = new Wallet(mnemonic);
            byte[] seed = wallet.DeriveMnemonicSeed();
            Account account = null;

            // 1. Encrypt mnemonic derived seed and generate keystore as json 
            string keystoreJson = keystoreService.EncryptAndGenerateDefaultKeyStoreAsJson(_securePassword, seed, wallet.Account.PublicKey);

            /* 2. Encrypt the mnemonic as bytes */
            byte[] stringByteArray = Encoding.UTF8.GetBytes(mnemonic.ToString());
            string encryptedKeystoreJson = keystoreService.EncryptAndGenerateDefaultKeyStoreAsJson(_securePassword, stringByteArray, wallet.Account.PublicKey.Key);

            string keystoreJsonAddr = SecretKeyStoreService.GetAddressFromKeyStore(encryptedKeystoreJson);

            /* 1. Decrypt mnemonic derived seed and generate wallet from it 
            var decryptedKeystore = keystoreService.DecryptKeyStoreFromJson(_password, keystoreJson);
            */

            /* 2. Decrypt the mnemonic as bytes */
            byte[] decryptedKeystore = keystoreService.DecryptKeyStoreFromJson(_securePassword, encryptedKeystoreJson);
            string mnemonicString = Encoding.UTF8.GetString(decryptedKeystore);

            /* 2. Restore the wallet from the restored mnemonic */
            Mnemonic restoredMnemonic = new Mnemonic(mnemonicString);
            Wallet restoredWallet = new Wallet(restoredMnemonic);

            // no passphrase to generate same keys as sollet
            //var restoredWallet = new Wallet.Wallet(decryptedKeystore);
            byte[] restoredSeed = restoredWallet.DeriveMnemonicSeed();

            // Mimic sollet key generation
            for (int idx = 0; idx < 10; idx++)
            {
                account = restoredWallet.GetAccount(idx);

                Assert.IsTrue(account.PublicKey.Key == expectedSolletAddresses[idx][0]);
                Assert.IsTrue(account.PrivateKey.Key == expectedSolletAddresses[idx][1]);
            }
        }

        [TestMethod]
        public void ApproveCheckedMultisigExample()
        {
            ApproveCheckedMultisigExample exampleClass = new ApproveCheckedMultisigExample();
            exampleClass.Run();
        }

        [TestMethod]
        public void HelloWorld()
        {
            var wallet = new Wallet(WordCount.TwentyFour, WordList.English);

            Console.WriteLine("Hello World!");
            Console.WriteLine($"Mnemonic: {wallet.Mnemonic}");
            Console.WriteLine($"PubKey: {wallet.Account.PublicKey.Key}");
            Console.WriteLine($"PrivateKey: {wallet.Account.PrivateKey.Key}");

            IRpcClient rpcClient = ClientFactory.GetClient(Cluster.TestNet);

            var balance = rpcClient.GetBalance(wallet.Account.PublicKey);

            Console.WriteLine($"Balance: {balance.Result.Value}");

            var transactionHash = rpcClient.RequestAirdrop(wallet.Account.PublicKey, 100_000_000);

            Console.WriteLine($"TxHash: {transactionHash.Result}");

            IStreamingRpcClient streamingRpcClient = ClientFactory.GetStreamingClient(Cluster.TestNet);

            streamingRpcClient.ConnectAsync().Wait();

            var subscription = streamingRpcClient.SubscribeSignature(transactionHash.Result, (sub, data) =>
            {
                if (data.Value.Error == null)
                {
                    var balance = rpcClient.GetBalance(wallet.Account.PublicKey);

                    Console.WriteLine($"Balance: {balance.Result.Value}");

                    var memoInstruction = MemoProgram.NewMemoV2("Hello Solana World, using Solnet :)");

                    var recentHash = rpcClient.GetRecentBlockHash();

                    var tx = new TransactionBuilder().AddInstruction(memoInstruction).SetFeePayer(wallet.Account)
                        .SetRecentBlockHash(recentHash.Result.Value.Blockhash).Build(wallet.Account);

                    var txHash = rpcClient.SendTransaction(tx);

                    Console.WriteLine($"TxHash: {txHash.Result}");
                }
                else
                {
                    Console.WriteLine($"Transaction error: {data.Value.Error.Type}");
                }
            });

            Console.ReadLine();
        }
    }
}