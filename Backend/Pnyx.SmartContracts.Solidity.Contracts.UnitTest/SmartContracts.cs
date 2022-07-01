using System.Numerics;
using Nethereum.Contracts;
using Nethereum.Contracts.CQS;
using Nethereum.Hex.HexTypes;
using Nethereum.Web3;
using Nethereum.Web3.Accounts;
using Nethereum.RPC.Eth.DTOs;
using Pnyx.SmartContracts.Solidity.Contracts.UnitTest.HelperClasses;

namespace Pnyx.SmartContracts.Solidity.Contracts.UnitTest
{
    [TestClass]
    public class SmartContracts
    {
		private const string DEFAULT_BYTECODE = "0x60606040526040516020806106f5833981016040528080519060200190919050505b80600160005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005081905550806000600050819055505b506106868061006f6000396000f360606040523615610074576000357c010000000000000000000000000000000000000000000000000000000090048063095ea7b31461008157806318160ddd146100b657806323b872dd146100d957806370a0823114610117578063a9059cbb14610143578063dd62ed3e1461017857610074565b61007f5b610002565b565b005b6100a060048080359060200190919080359060200190919050506101ad565b6040518082815260200191505060405180910390f35b6100c36004805050610674565b6040518082815260200191505060405180910390f35b6101016004808035906020019091908035906020019091908035906020019091905050610281565b6040518082815260200191505060405180910390f35b61012d600480803590602001909190505061048d565b6040518082815260200191505060405180910390f35b61016260048080359060200190919080359060200190919050506104cb565b6040518082815260200191505060405180910390f35b610197600480803590602001909190803590602001909190505061060b565b6040518082815260200191505060405180910390f35b600081600260005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060008573ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040518082815260200191505060405180910390a36001905061027b565b92915050565b600081600160005060008673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050541015801561031b575081600260005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060003373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505410155b80156103275750600082115b1561047c5781600160005060008573ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a381600160005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555081600260005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060003373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505403925050819055506001905061048656610485565b60009050610486565b5b9392505050565b6000600160005060008373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505490506104c6565b919050565b600081600160005060003373ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050541015801561050c5750600082115b156105fb5781600160005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555081600160005060008573ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a36001905061060556610604565b60009050610605565b5b92915050565b6000600260005060008473ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060008373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054905061066e565b92915050565b60006000600050549050610683565b9056";

		[TestMethod]
        [TestCategory("IsDeployable")]
        public void IsDeployable_PnyxGovernor()
        {
			string contractAddress = String.Empty;

			//  Instantiating Web3 and the Account

			// To create an instance of web3 we first provide the url of our testchain and the private key of our account. 
			// When providing an Account instantiated with a  private key all our transactions will be signed “offline” by Nethereum.

			string bytecode = BytecodeProvider.Load("Governors", "PnyxGovernor");
			string privateKey = Pnyx_SmartContracts_Solidity_Contracts.Default.DefaultPrivateKey;
			BigInteger chainId = Pnyx_SmartContracts_Solidity_Contracts.Default.DefaultChainId;
			Account account = new Account(privateKey, chainId);
			Console.WriteLine("Our account: " + account.Address);

			//Now let's create an instance of Web3 using our account pointing to our nethereum testchain
			Web3 web3 = new Web3(account, "http://testchain.nethereum.com:8545");
			// web3.TransactionManager.UseLegacyAsDefault = true; //Using legacy option instead of 1559

			//  Deploying the Contract
			// The next step is to deploy our Standard Token ERC20 smart contract, in this scenario the total supply (number of tokens) is going to be 100,000.
			// First we create an instance of the StandardTokenDeployment with the TotalSupply amount.

			StandardTokenDeployment deploymentMessage = new StandardTokenDeployment(bytecode)
			{
				//TotalSupply = 100000,
				Gas = 10000000
			};

			// Then we create a deployment handler using our contract deployment definition and simply deploy the contract using the deployment message. We are auto estimating the gas, getting the latest gas price and nonce so nothing else is set anything on the deployment message.
			// Finally, we wait for the deployment transaction to be mined, and retrieve the contract address of the new contract from the receipt.

			IContractDeploymentTransactionHandler<StandardTokenDeployment> deploymentHandler = web3.Eth.GetContractDeploymentHandler<StandardTokenDeployment>();

			//HexBigInteger estimate = deploymentHandler.EstimateGasAsync(deploymentMessage).Result;
			//BigInteger gasEstimate = estimate.Value;

			try
			{
				TransactionReceipt transactionReceipt = deploymentHandler.SendRequestAndWaitForReceiptAsync(deploymentMessage).Result;

				contractAddress = transactionReceipt.ContractAddress;

				Console.WriteLine("Deployed Contract address is: " + contractAddress);
			}
			catch (Exception ex)
            {
				Console.WriteLine(String.Format("IsDeployable_PnyxGovernor failed with message \"{0}\"", ex.Message));
				Assert.Fail(ex.Message);
			}

			Assert.IsFalse(String.IsNullOrEmpty(contractAddress));
		}
	}
}