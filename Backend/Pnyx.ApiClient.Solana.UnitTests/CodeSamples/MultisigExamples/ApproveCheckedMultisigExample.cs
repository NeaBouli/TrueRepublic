using Solnet.Programs;
using Solnet.Rpc;
using Solnet.Rpc.Builders;
using Solnet.Rpc.Core.Http;
using Solnet.Rpc.Messages;
using Solnet.Rpc.Models;
using Solnet.Wallet;
using System;
using System.Collections.Generic;

namespace Pnyx.ApiClient.Solana.UnitTests.CodeSamples.MultisigExamples
{
    /// <summary>
    /// Example of how to approve and revoke a delegate to transfer tokens using multisig
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class ApproveCheckedMultisigExample
    {
        private static readonly IRpcClient rpcClient = ClientFactory.GetClient(Cluster.TestNet);

        private const string MnemonicWords =
            "route clerk disease box emerge airport loud waste attitude film army tray " +
            "forward deal onion eight catalog surface unit card window walnut wealth medal";

        public void Run()
        {
            Wallet wallet = new(MnemonicWords);

            RequestResult<ResponseValue<BlockHash>> blockHash = rpcClient.GetRecentBlockHash();

            ulong minBalanceForExemptionMultiSig =
                rpcClient.GetMinimumBalanceForRentExemption(TokenProgram.MultisigAccountDataSize).Result;
            Console.WriteLine($"MinBalanceForRentExemption MultiSig >> {minBalanceForExemptionMultiSig}");
            ulong minBalanceForExemptionAcc =
                rpcClient.GetMinimumBalanceForRentExemption(TokenProgram.TokenAccountDataSize).Result;
            Console.WriteLine($"MinBalanceForRentExemption Account >> {minBalanceForExemptionAcc}");
            ulong minBalanceForExemptionMint =
                rpcClient.GetMinimumBalanceForRentExemption(TokenProgram.MintAccountDataSize).Result;
            Console.WriteLine($"MinBalanceForRentExemption Mint Account >> {minBalanceForExemptionMint}");

            Account ownerAccount = wallet.GetAccount(10);
            Account delegateAccount = wallet.GetAccount(194330);
            Account mintAccount = wallet.GetAccount(94330);
            Account initialAccount = wallet.GetAccount(84330);

            // the token mint multisig
            Account mintMultiSignature = wallet.GetAccount(10116);

            // the signers for the token mint authority multisig
            Account mintSigner1 = wallet.GetAccount(251280);
            Account mintSigner2 = wallet.GetAccount(251281);
            Account mintSigner3 = wallet.GetAccount(251282);
            Account mintSigner4 = wallet.GetAccount(251283);
            Account mintSigner5 = wallet.GetAccount(251284);

            // The token account
            Account tokenAccountWithMultisigOwner = wallet.GetAccount(4044);
            // The multisig which is the token account authority
            Account tokenMultiSignature = wallet.GetAccount(4045);

            // the signers for the token authority multisig
            Account tokenAccountSigner1 = wallet.GetAccount(25490);
            Account tokenAccountSigner2 = wallet.GetAccount(25491);
            Account tokenAccountSigner3 = wallet.GetAccount(25492);
            Account tokenAccountSigner4 = wallet.GetAccount(25493);
            Account tokenAccountSigner5 = wallet.GetAccount(25494);

            byte[] msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount.PublicKey,
                    tokenMultiSignature,
                    minBalanceForExemptionMultiSig,
                    TokenProgram.MultisigAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeMultiSignature(
                    tokenMultiSignature,
                    new List<PublicKey>
                    {
                        tokenAccountSigner1,
                        tokenAccountSigner2,
                        tokenAccountSigner3,
                        tokenAccountSigner4,
                        tokenAccountSigner5
                    },
                    3))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            Message msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            Transaction tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenMultiSignature.Sign(msgData),
                });

            byte[] txBytes = Examples.LogTransactionAndSerialize(tx);

            string signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();

            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount.PublicKey,
                    tokenAccountWithMultisigOwner,
                    minBalanceForExemptionAcc,
                    TokenProgram.TokenAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeAccount(
                    tokenAccountWithMultisigOwner,
                    mintAccount,
                    tokenMultiSignature))
                .AddInstruction(TokenProgram.MintTo(
                    mintAccount,
                    tokenAccountWithMultisigOwner,
                    25000,
                    mintMultiSignature,
                    new List<PublicKey>
                    {
                        mintSigner1,
                        mintSigner2,
                        mintSigner4
                    }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenAccountWithMultisigOwner.Sign(msgData),
                    mintSigner1.Sign(msgData),
                    mintSigner2.Sign(msgData),
                    mintSigner4.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();

            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.ApproveChecked(
                        tokenAccountWithMultisigOwner,
                        delegateAccount,
                        5000,
                        10,
                        tokenMultiSignature,
                        mintAccount,
                        new List<PublicKey>
                        {
                            tokenAccountSigner1,
                            tokenAccountSigner2,
                            tokenAccountSigner3,
                        }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenAccountSigner1.Sign(msgData),
                    tokenAccountSigner2.Sign(msgData),
                    tokenAccountSigner3.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();


            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.TransferChecked(
                    tokenAccountWithMultisigOwner,
                    initialAccount,
                    5000,
                    10,
                    delegateAccount,
                    mintAccount))
                .AddInstruction(TokenProgram.Revoke(
                    tokenAccountWithMultisigOwner,
                    tokenMultiSignature,
                    new List<PublicKey>
                    {
                        tokenAccountSigner1,
                        tokenAccountSigner2,
                        tokenAccountSigner3,
                    }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]> {
                    ownerAccount.Sign(msgData),
                    delegateAccount.Sign(msgData),
                    tokenAccountSigner1.Sign(msgData),
                    tokenAccountSigner2.Sign(msgData),
                    tokenAccountSigner3.Sign(msgData)
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

        }
    }
}
