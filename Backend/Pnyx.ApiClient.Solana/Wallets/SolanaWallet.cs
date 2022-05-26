using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.Serialization;
using System.Text;
using System.Threading.Tasks;
using Solnet.Wallet;
using Solnet.Wallet.Bip39;

namespace Pnyx.ApiClient.Solana.Wallets
{
    public class SolanaWallet : ISerializable
    {
        #region Statics

        public static SolanaWallet CreateNew()
        {
            Wallet wallet = new Wallet(WordCount.TwentyFour, WordList.English);
            return new SolanaWallet(wallet);
        }

        public static SolanaWallet Open()
        {
            return null;
        }

        #endregion

        #region Class Members

        Wallet _wallet = null;

        #endregion

        #region Construction / Initialization

        public SolanaWallet(Wallet newWallet)
        {
            _wallet = new Wallet(WordCount.TwentyFour, WordList.English);
        }

        #endregion

        #region Properties


        #endregion

        #region ISerializable

        public void GetObjectData(SerializationInfo info, StreamingContext context)
        {
            throw new NotImplementedException();
        }

        #endregion
    }
}
