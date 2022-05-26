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
    /// Example of how to close a multisig account
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class BurnCheckedAndCloseAccountMultisigExample
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
            Account mintAccount = wallet.GetAccount(94330);

            // The multisig which is the token account authority
            Account tokenAccountWithMultisigOwner = wallet.GetAccount(4044);
            Account tokenMultiSignature = wallet.GetAccount(4045);
            Account tokenAccountSigner1 = wallet.GetAccount(25490);
            Account tokenAccountSigner2 = wallet.GetAccount(25491);
            Account tokenAccountSigner3 = wallet.GetAccount(25492);

            // The account has balance so we'll burn it before
            RequestResult<ResponseValue<TokenBalance>> balance =
                rpcClient.GetTokenAccountBalance(tokenAccountWithMultisigOwner.PublicKey);

            Console.WriteLine($"Account Balance >> {balance.Result.Value.UiAmountString}");

            byte[] msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.BurnChecked(
                    mintAccount,
                    tokenAccountWithMultisigOwner,
                    tokenMultiSignature,
                    balance.Result.Value.AmountUlong,
                    10,
                    new List<PublicKey>()
                    {
                        tokenAccountSigner1,
                        tokenAccountSigner2,
                        tokenAccountSigner3
                    }))
                .AddInstruction(TokenProgram.CloseAccount(
                    tokenAccountWithMultisigOwner,
                    ownerAccount,
                    tokenMultiSignature,
                    TokenProgram.ProgramIdKey,
                    new List<PublicKey>()
                    {
                        tokenAccountSigner1,
                        tokenAccountSigner2,
                        tokenAccountSigner3
                    }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            Message msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            Transaction tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    tokenAccountSigner1.Sign(msgData),
                    tokenAccountSigner2.Sign(msgData),
                    tokenAccountSigner3.Sign(msgData),
                });

            byte[] txBytes = Examples.LogTransactionAndSerialize(tx);

            string signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

        }
    }
}
