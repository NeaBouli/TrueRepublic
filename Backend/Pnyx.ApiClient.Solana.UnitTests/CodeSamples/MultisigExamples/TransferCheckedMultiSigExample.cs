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
    /// An example on how to use multisig accounts to control a token account.
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class TransferCheckedMultiSigExample
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
            Account mintAccount = wallet.GetAccount(94224);
            Account initialAccount = wallet.GetAccount(84224);

            Account tokenAccountWithMultisigOwner = wallet.GetAccount(3042);
            Account tokenMultiSignature = wallet.GetAccount(3043);

            // The signers for the token account
            Account tokenAccountSigner1 = wallet.GetAccount(25280);
            Account tokenAccountSigner2 = wallet.GetAccount(25281);
            Account tokenAccountSigner3 = wallet.GetAccount(25282);
            Account tokenAccountSigner4 = wallet.GetAccount(25283);
            Account tokenAccountSigner5 = wallet.GetAccount(25284);

            // First we create a multi sig account to use as the token account authority
            // In this same transaction we transfer tokens using TransferChecked from the initialAccount in the example above
            // to the same token account we just finished creating
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
                .AddInstruction(TokenProgram.TransferChecked(
                    initialAccount,
                    tokenAccountWithMultisigOwner,
                    10000, 10,
                    ownerAccount,
                    mintAccount))
                .CompileMessage();

            Message msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            Transaction tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenMultiSignature.Sign(msgData),
                    tokenAccountWithMultisigOwner.Sign(msgData),
                });

            byte[] txBytes = Examples.LogTransactionAndSerialize(tx);

            string signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            // After the previous transaction is confirmed we use TransferChecked to transfer tokens using the
            // multi sig account back to the initial account
            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.Transfer(
                    tokenAccountWithMultisigOwner,
                    initialAccount,
                    10000,
                    tokenMultiSignature,
                    new List<PublicKey>()
                    {
                        tokenAccountSigner3,
                        tokenAccountSigner4,
                        tokenAccountSigner5
                    })).CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenAccountSigner3.Sign(msgData),
                    tokenAccountSigner4.Sign(msgData),
                    tokenAccountSigner5.Sign(msgData)
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);
        }
    }
}
