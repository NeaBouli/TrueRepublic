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
    /// An example on how to control the freeze authority of a token using multi signatures
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class FreezeAuthorityExample
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

            // the signers for the token mint
            Account mintMultiSignature = wallet.GetAccount(10116);
            Account mintSigner1 = wallet.GetAccount(251280);
            Account mintSigner2 = wallet.GetAccount(251281);
            Account mintSigner3 = wallet.GetAccount(251282);
            Account mintSigner4 = wallet.GetAccount(251283);
            Account mintSigner5 = wallet.GetAccount(251284);

            // The signers for the freeze account
            Account freezeMultiSignature = wallet.GetAccount(3057);
            Account freezeSigner1 = wallet.GetAccount(25410);
            Account freezeSigner2 = wallet.GetAccount(25411);
            Account freezeSigner3 = wallet.GetAccount(25412);
            Account freezeSigner4 = wallet.GetAccount(25413);
            Account freezeSigner5 = wallet.GetAccount(25414);

            // First we create a multi sig account to use as the token's freeze authority
            byte[] msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount,
                    freezeMultiSignature,
                    minBalanceForExemptionMultiSig,
                    TokenProgram.MultisigAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeMultiSignature(
                    freezeMultiSignature,
                    new List<PublicKey>
                    {
                        freezeSigner1,
                        freezeSigner2,
                        freezeSigner3,
                        freezeSigner4,
                        freezeSigner5
                    }, 3))
                .CompileMessage();

            Message msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            Transaction tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    freezeMultiSignature.Sign(msgData),
                });

            byte[] txBytes = Examples.LogTransactionAndSerialize(tx);

            string signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();


            // Then we create an account which will be the token's mint authority
            // In this same transaction we initialize the token mint with said authorities
            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount,
                    mintMultiSignature,
                    minBalanceForExemptionMultiSig,
                    TokenProgram.MultisigAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeMultiSignature(
                    mintMultiSignature,
                    new List<PublicKey>
                    {
                        mintSigner1,
                        mintSigner2,
                        mintSigner3,
                        mintSigner4,
                        mintSigner5
                    }, 3))
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount,
                    mintAccount,
                    minBalanceForExemptionMint,
                    TokenProgram.MintAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeMint(
                    mintAccount,
                    10,
                    mintMultiSignature,
                    freezeMultiSignature))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    mintMultiSignature.Sign(msgData),
                    mintAccount.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();

            // Here we mint tokens to an account using the mint authority multi sig
            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(SystemProgram.CreateAccount(
                    ownerAccount,
                    initialAccount,
                    minBalanceForExemptionAcc,
                    TokenProgram.TokenAccountDataSize,
                    TokenProgram.ProgramIdKey))
                .AddInstruction(TokenProgram.InitializeAccount(
                    initialAccount,
                    mintAccount,
                    ownerAccount.PublicKey))
                .AddInstruction(TokenProgram.MintTo(
                    mintAccount,
                    initialAccount,
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
                    initialAccount.Sign(msgData),
                    mintSigner1.Sign(msgData),
                    mintSigner2.Sign(msgData),
                    mintSigner4.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();

            // After doing this, we freeze the account to which we just minted tokens
            // Notice how the signers used are different, because the `freezeAuthority` has different signers
            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.FreezeAccount(
                        initialAccount,
                        mintAccount,
                        freezeMultiSignature,
                        TokenProgram.ProgramIdKey,
                        new List<PublicKey>
                        {
                            freezeSigner2,
                            freezeSigner3,
                            freezeSigner4,
                        }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    freezeSigner2.Sign(msgData),
                    freezeSigner3.Sign(msgData),
                    freezeSigner4.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);

            blockHash = rpcClient.GetRecentBlockHash();

            // Because we're actually cool people, we now thaw that same account and then set the authority to nothing
            msgData = new TransactionBuilder().SetRecentBlockHash(blockHash.Result.Value.Blockhash)
                .SetFeePayer(ownerAccount)
                .AddInstruction(TokenProgram.ThawAccount(
                    initialAccount,
                    mintAccount,
                    freezeMultiSignature,
                    TokenProgram.ProgramIdKey,
                    new List<PublicKey>
                    {
                        freezeSigner2,
                        freezeSigner3,
                        freezeSigner4,
                    }))
                .AddInstruction(TokenProgram.SetAuthority(
                    mintAccount,
                    AuthorityType.FreezeAccount,
                    freezeMultiSignature,
                    null,
                    new List<PublicKey>
                    {
                        freezeSigner2,
                        freezeSigner3,
                        freezeSigner4,
                    }))
                .AddInstruction(MemoProgram.NewMemo(ownerAccount, "Hello from Sol.Net"))
                .CompileMessage();

            msg = Examples.DecodeMessageFromWire(msgData);

            Console.WriteLine("\n\tPOPULATING TRANSACTION WITH SIGNATURES\t");
            tx = Transaction.Populate(msg,
                new List<byte[]>
                {
                    ownerAccount.Sign(msgData),
                    freezeSigner2.Sign(msgData),
                    freezeSigner3.Sign(msgData),
                    freezeSigner4.Sign(msgData),
                });

            txBytes = Examples.LogTransactionAndSerialize(tx);

            signature = Examples.SubmitTxSendAndLog(txBytes);
            Examples.PollConfirmedTx(signature);
        }
    }
}
