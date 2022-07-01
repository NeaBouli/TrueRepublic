using Nethereum.Web3;
using Nethereum.ABI.FunctionEncoding.Attributes;
using Nethereum.Contracts.CQS;
using Nethereum.Util;
using Nethereum.Web3.Accounts;
using Nethereum.Hex.HexConvertors.Extensions;
using Nethereum.Contracts;
using Nethereum.Contracts.Extensions;
using System;
using System.Numerics;
using System.Threading;
using System.Threading.Tasks;

namespace Pnyx.SmartContracts.Solidity.Contracts.UnitTest.HelperClasses
{
	internal class StandardTokenDeployment : ContractDeploymentMessage
	{
		public StandardTokenDeployment()
			: base(String.Empty)
		{
		}

		public StandardTokenDeployment(string byteCode)
			: base(byteCode)
		{
		}

		[Parameter("uint256", "totalSupply")]
		public BigInteger TotalSupply { get; set; }
	}
}
