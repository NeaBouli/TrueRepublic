using Solnet.Programs.Models.TokenProgram;
using Solnet.Rpc;
using Solnet.Wallet;
using System;

namespace Pnyx.ApiClient.Solana.UnitTests.CodeSamples.MultisigExamples
{
    /// <summary>
    /// Code from https://github.com/bmresearch/Solnet/blob/master/src/Solnet.Examples/MultisigExamples.cs
    /// </summary>
    public class GetMultiSignatureAccountExample
    {
        private static readonly IRpcClient rpcClient = ClientFactory.GetClient(Cluster.TestNet);

        private const string MnemonicWords =
            "route clerk disease box emerge airport loud waste attitude film army tray " +
            "forward deal onion eight catalog surface unit card window walnut wealth medal";


        public void Run()
        {
            Wallet wallet = new(MnemonicWords);

            // The multisig which is the token account authority
            Account tokenMultiSignature = wallet.GetAccount(4045);

            var account = rpcClient.GetAccountInfo(tokenMultiSignature.PublicKey);

            var multiSigAccount = MultiSignatureAccount.Deserialize(Convert.FromBase64String(account.Result.Value.Data[0]));


        }
    }
}
