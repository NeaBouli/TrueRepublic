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
    /// Example of how to mint and burn using multisigs
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class SimpleMintToAndBurnCheckedMultisigExample
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
            Account initialAccount = wallet.GetAccount(84330);

            // the token mint multisig
            Account mintMultiSignature = wallet.GetAccount(10116);
            Account mintSigner1 = wallet.GetAccount(251280);
            Account mintSigner2 = wallet.GetAccount(251281);
            Account mintSigner3 = wallet.GetAccount(251282);

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
                .AddInstruction(TokenProgram.MintToChecked(
                    mintAccount,
                    tokenAccountWithMultisigOwner,
                    mintMultiSignature,
                    1_000_000_000,
                    10,
                    new List<PublicKey>()
                    {
                        mintSigner1,
                        mintSigner2,
                        mintSigner3
                    }))
                .AddInstruction(TokenProgram.BurnChecked(
                    mintAccount,
                    tokenAccountWithMultisigOwner,
                    tokenMultiSignature,
                    500_000,
                    10,
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
                    mintSigner1.Sign(msgData),
                    mintSigner2.Sign(msgData),
                    mintSigner3.Sign(msgData),
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
