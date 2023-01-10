using System;
using System.Collections.Generic;
using System.Numerics;
using System.Text;
using Pnyx.ApiClient.Core.Accounts;
using Nethereum.Web3.Accounts;
using Nethereum.Signer;

namespace Pnyx.SmartContracts.Solidity.Contracts.Accounts
{
    public class ETHPnyxAccount : PnyxAccount
    {
        public static ETHPnyxAccount Create(string privateKey, BigInteger? chainId = null)
        {
            return new ETHPnyxAccount(privateKey, chainId);
        }

        public ETHPnyxAccount(string privateKey, BigInteger? chainId = null)
            : base()
        {
            ETHAccount = new Account(privateKey, chainId);
        }

        public Account ETHAccount { get; }

        public PnyxAccountRoles AccountRole { get; }
    }
}
