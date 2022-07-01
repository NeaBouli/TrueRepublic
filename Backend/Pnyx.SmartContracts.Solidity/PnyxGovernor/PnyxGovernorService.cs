using System;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Numerics;
using Nethereum.Hex.HexTypes;
using Nethereum.ABI.FunctionEncoding.Attributes;
using Nethereum.Web3;
using Nethereum.RPC.Eth.DTOs;
using Nethereum.Contracts.CQS;
using Nethereum.Contracts.ContractHandlers;
using Nethereum.Contracts;
using System.Threading;
using Pnyx.SmartContracts.Solidity.Contracts.PnyxGovernor.ContractDefinition;

namespace Pnyx.SmartContracts.Solidity.Contracts.PnyxGovernor
{
    public partial class PnyxGovernorService
    {
        public static Task<TransactionReceipt> DeployContractAndWaitForReceiptAsync(Nethereum.Web3.Web3 web3, PnyxGovernorDeployment pnyxGovernorDeployment, CancellationTokenSource cancellationTokenSource = null)
        {
            return web3.Eth.GetContractDeploymentHandler<PnyxGovernorDeployment>().SendRequestAndWaitForReceiptAsync(pnyxGovernorDeployment, cancellationTokenSource);
        }

        public static Task<string> DeployContractAsync(Nethereum.Web3.Web3 web3, PnyxGovernorDeployment pnyxGovernorDeployment)
        {
            return web3.Eth.GetContractDeploymentHandler<PnyxGovernorDeployment>().SendRequestAsync(pnyxGovernorDeployment);
        }

        public static async Task<PnyxGovernorService> DeployContractAndGetServiceAsync(Nethereum.Web3.Web3 web3, PnyxGovernorDeployment pnyxGovernorDeployment, CancellationTokenSource cancellationTokenSource = null)
        {
            var receipt = await DeployContractAndWaitForReceiptAsync(web3, pnyxGovernorDeployment, cancellationTokenSource);
            return new PnyxGovernorService(web3, receipt.ContractAddress);
        }

        protected Nethereum.Web3.Web3 Web3{ get; }

        public ContractHandler ContractHandler { get; }

        public PnyxGovernorService(Nethereum.Web3.Web3 web3, string contractAddress)
        {
            Web3 = web3;
            ContractHandler = web3.Eth.GetContractHandler(contractAddress);
        }

        public Task<byte[]> BallotTypehashQueryAsync(BallotTypehashFunction ballotTypehashFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<BallotTypehashFunction, byte[]>(ballotTypehashFunction, blockParameter);
        }

        
        public Task<byte[]> BallotTypehashQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<BallotTypehashFunction, byte[]>(null, blockParameter);
        }

        public Task<string> CountingModeQueryAsync(CountingModeFunction countingModeFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<CountingModeFunction, string>(countingModeFunction, blockParameter);
        }

        
        public Task<string> CountingModeQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<CountingModeFunction, string>(null, blockParameter);
        }

        public Task<byte[]> ExtendedBallotTypehashQueryAsync(ExtendedBallotTypehashFunction extendedBallotTypehashFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ExtendedBallotTypehashFunction, byte[]>(extendedBallotTypehashFunction, blockParameter);
        }

        
        public Task<byte[]> ExtendedBallotTypehashQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ExtendedBallotTypehashFunction, byte[]>(null, blockParameter);
        }

        public Task<string> CastVoteRequestAsync(CastVoteFunction castVoteFunction)
        {
             return ContractHandler.SendRequestAsync(castVoteFunction);
        }

        public Task<TransactionReceipt> CastVoteRequestAndWaitForReceiptAsync(CastVoteFunction castVoteFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteFunction, cancellationToken);
        }

        public Task<string> CastVoteRequestAsync(BigInteger proposalId, byte support)
        {
            var castVoteFunction = new CastVoteFunction();
                castVoteFunction.ProposalId = proposalId;
                castVoteFunction.Support = support;
            
             return ContractHandler.SendRequestAsync(castVoteFunction);
        }

        public Task<TransactionReceipt> CastVoteRequestAndWaitForReceiptAsync(BigInteger proposalId, byte support, CancellationTokenSource cancellationToken = null)
        {
            var castVoteFunction = new CastVoteFunction();
                castVoteFunction.ProposalId = proposalId;
                castVoteFunction.Support = support;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteFunction, cancellationToken);
        }

        public Task<string> CastVoteBySigRequestAsync(CastVoteBySigFunction castVoteBySigFunction)
        {
             return ContractHandler.SendRequestAsync(castVoteBySigFunction);
        }

        public Task<TransactionReceipt> CastVoteBySigRequestAndWaitForReceiptAsync(CastVoteBySigFunction castVoteBySigFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteBySigFunction, cancellationToken);
        }

        public Task<string> CastVoteBySigRequestAsync(BigInteger proposalId, byte support, byte v, byte[] r, byte[] s)
        {
            var castVoteBySigFunction = new CastVoteBySigFunction();
                castVoteBySigFunction.ProposalId = proposalId;
                castVoteBySigFunction.Support = support;
                castVoteBySigFunction.V = v;
                castVoteBySigFunction.R = r;
                castVoteBySigFunction.S = s;
            
             return ContractHandler.SendRequestAsync(castVoteBySigFunction);
        }

        public Task<TransactionReceipt> CastVoteBySigRequestAndWaitForReceiptAsync(BigInteger proposalId, byte support, byte v, byte[] r, byte[] s, CancellationTokenSource cancellationToken = null)
        {
            var castVoteBySigFunction = new CastVoteBySigFunction();
                castVoteBySigFunction.ProposalId = proposalId;
                castVoteBySigFunction.Support = support;
                castVoteBySigFunction.V = v;
                castVoteBySigFunction.R = r;
                castVoteBySigFunction.S = s;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteBySigFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonRequestAsync(CastVoteWithReasonFunction castVoteWithReasonFunction)
        {
             return ContractHandler.SendRequestAsync(castVoteWithReasonFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonRequestAndWaitForReceiptAsync(CastVoteWithReasonFunction castVoteWithReasonFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonRequestAsync(BigInteger proposalId, byte support, string reason)
        {
            var castVoteWithReasonFunction = new CastVoteWithReasonFunction();
                castVoteWithReasonFunction.ProposalId = proposalId;
                castVoteWithReasonFunction.Support = support;
                castVoteWithReasonFunction.Reason = reason;
            
             return ContractHandler.SendRequestAsync(castVoteWithReasonFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonRequestAndWaitForReceiptAsync(BigInteger proposalId, byte support, string reason, CancellationTokenSource cancellationToken = null)
        {
            var castVoteWithReasonFunction = new CastVoteWithReasonFunction();
                castVoteWithReasonFunction.ProposalId = proposalId;
                castVoteWithReasonFunction.Support = support;
                castVoteWithReasonFunction.Reason = reason;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonAndParamsRequestAsync(CastVoteWithReasonAndParamsFunction castVoteWithReasonAndParamsFunction)
        {
             return ContractHandler.SendRequestAsync(castVoteWithReasonAndParamsFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonAndParamsRequestAndWaitForReceiptAsync(CastVoteWithReasonAndParamsFunction castVoteWithReasonAndParamsFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonAndParamsFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonAndParamsRequestAsync(BigInteger proposalId, byte support, string reason, byte[] @params)
        {
            var castVoteWithReasonAndParamsFunction = new CastVoteWithReasonAndParamsFunction();
                castVoteWithReasonAndParamsFunction.ProposalId = proposalId;
                castVoteWithReasonAndParamsFunction.Support = support;
                castVoteWithReasonAndParamsFunction.Reason = reason;
                castVoteWithReasonAndParamsFunction.Params = @params;
            
             return ContractHandler.SendRequestAsync(castVoteWithReasonAndParamsFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonAndParamsRequestAndWaitForReceiptAsync(BigInteger proposalId, byte support, string reason, byte[] @params, CancellationTokenSource cancellationToken = null)
        {
            var castVoteWithReasonAndParamsFunction = new CastVoteWithReasonAndParamsFunction();
                castVoteWithReasonAndParamsFunction.ProposalId = proposalId;
                castVoteWithReasonAndParamsFunction.Support = support;
                castVoteWithReasonAndParamsFunction.Reason = reason;
                castVoteWithReasonAndParamsFunction.Params = @params;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonAndParamsFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonAndParamsBySigRequestAsync(CastVoteWithReasonAndParamsBySigFunction castVoteWithReasonAndParamsBySigFunction)
        {
             return ContractHandler.SendRequestAsync(castVoteWithReasonAndParamsBySigFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonAndParamsBySigRequestAndWaitForReceiptAsync(CastVoteWithReasonAndParamsBySigFunction castVoteWithReasonAndParamsBySigFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonAndParamsBySigFunction, cancellationToken);
        }

        public Task<string> CastVoteWithReasonAndParamsBySigRequestAsync(BigInteger proposalId, byte support, string reason, byte[] @params, byte v, byte[] r, byte[] s)
        {
            var castVoteWithReasonAndParamsBySigFunction = new CastVoteWithReasonAndParamsBySigFunction();
                castVoteWithReasonAndParamsBySigFunction.ProposalId = proposalId;
                castVoteWithReasonAndParamsBySigFunction.Support = support;
                castVoteWithReasonAndParamsBySigFunction.Reason = reason;
                castVoteWithReasonAndParamsBySigFunction.Params = @params;
                castVoteWithReasonAndParamsBySigFunction.V = v;
                castVoteWithReasonAndParamsBySigFunction.R = r;
                castVoteWithReasonAndParamsBySigFunction.S = s;
            
             return ContractHandler.SendRequestAsync(castVoteWithReasonAndParamsBySigFunction);
        }

        public Task<TransactionReceipt> CastVoteWithReasonAndParamsBySigRequestAndWaitForReceiptAsync(BigInteger proposalId, byte support, string reason, byte[] @params, byte v, byte[] r, byte[] s, CancellationTokenSource cancellationToken = null)
        {
            var castVoteWithReasonAndParamsBySigFunction = new CastVoteWithReasonAndParamsBySigFunction();
                castVoteWithReasonAndParamsBySigFunction.ProposalId = proposalId;
                castVoteWithReasonAndParamsBySigFunction.Support = support;
                castVoteWithReasonAndParamsBySigFunction.Reason = reason;
                castVoteWithReasonAndParamsBySigFunction.Params = @params;
                castVoteWithReasonAndParamsBySigFunction.V = v;
                castVoteWithReasonAndParamsBySigFunction.R = r;
                castVoteWithReasonAndParamsBySigFunction.S = s;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(castVoteWithReasonAndParamsBySigFunction, cancellationToken);
        }

        public Task<string> ExecuteRequestAsync(ExecuteFunction executeFunction)
        {
             return ContractHandler.SendRequestAsync(executeFunction);
        }

        public Task<TransactionReceipt> ExecuteRequestAndWaitForReceiptAsync(ExecuteFunction executeFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(executeFunction, cancellationToken);
        }

        public Task<string> ExecuteRequestAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, byte[] descriptionHash)
        {
            var executeFunction = new ExecuteFunction();
                executeFunction.Targets = targets;
                executeFunction.Values = values;
                executeFunction.Calldatas = calldatas;
                executeFunction.DescriptionHash = descriptionHash;
            
             return ContractHandler.SendRequestAsync(executeFunction);
        }

        public Task<TransactionReceipt> ExecuteRequestAndWaitForReceiptAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, byte[] descriptionHash, CancellationTokenSource cancellationToken = null)
        {
            var executeFunction = new ExecuteFunction();
                executeFunction.Targets = targets;
                executeFunction.Values = values;
                executeFunction.Calldatas = calldatas;
                executeFunction.DescriptionHash = descriptionHash;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(executeFunction, cancellationToken);
        }

        public Task<BigInteger> GetVotesQueryAsync(GetVotesFunction getVotesFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<GetVotesFunction, BigInteger>(getVotesFunction, blockParameter);
        }

        
        public Task<BigInteger> GetVotesQueryAsync(string account, BigInteger blockNumber, BlockParameter blockParameter = null)
        {
            var getVotesFunction = new GetVotesFunction();
                getVotesFunction.Account = account;
                getVotesFunction.BlockNumber = blockNumber;
            
            return ContractHandler.QueryAsync<GetVotesFunction, BigInteger>(getVotesFunction, blockParameter);
        }

        public Task<BigInteger> GetVotesWithParamsQueryAsync(GetVotesWithParamsFunction getVotesWithParamsFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<GetVotesWithParamsFunction, BigInteger>(getVotesWithParamsFunction, blockParameter);
        }

        
        public Task<BigInteger> GetVotesWithParamsQueryAsync(string account, BigInteger blockNumber, byte[] @params, BlockParameter blockParameter = null)
        {
            var getVotesWithParamsFunction = new GetVotesWithParamsFunction();
                getVotesWithParamsFunction.Account = account;
                getVotesWithParamsFunction.BlockNumber = blockNumber;
                getVotesWithParamsFunction.Params = @params;
            
            return ContractHandler.QueryAsync<GetVotesWithParamsFunction, BigInteger>(getVotesWithParamsFunction, blockParameter);
        }

        public Task<bool> HasVotedQueryAsync(HasVotedFunction hasVotedFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<HasVotedFunction, bool>(hasVotedFunction, blockParameter);
        }

        
        public Task<bool> HasVotedQueryAsync(BigInteger proposalId, string account, BlockParameter blockParameter = null)
        {
            var hasVotedFunction = new HasVotedFunction();
                hasVotedFunction.ProposalId = proposalId;
                hasVotedFunction.Account = account;
            
            return ContractHandler.QueryAsync<HasVotedFunction, bool>(hasVotedFunction, blockParameter);
        }

        public Task<BigInteger> HashProposalQueryAsync(HashProposalFunction hashProposalFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<HashProposalFunction, BigInteger>(hashProposalFunction, blockParameter);
        }

        
        public Task<BigInteger> HashProposalQueryAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, byte[] descriptionHash, BlockParameter blockParameter = null)
        {
            var hashProposalFunction = new HashProposalFunction();
                hashProposalFunction.Targets = targets;
                hashProposalFunction.Values = values;
                hashProposalFunction.Calldatas = calldatas;
                hashProposalFunction.DescriptionHash = descriptionHash;
            
            return ContractHandler.QueryAsync<HashProposalFunction, BigInteger>(hashProposalFunction, blockParameter);
        }

        public Task<string> NameQueryAsync(NameFunction nameFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<NameFunction, string>(nameFunction, blockParameter);
        }

        
        public Task<string> NameQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<NameFunction, string>(null, blockParameter);
        }

        public Task<string> OnERC1155BatchReceivedRequestAsync(OnERC1155BatchReceivedFunction onERC1155BatchReceivedFunction)
        {
             return ContractHandler.SendRequestAsync(onERC1155BatchReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC1155BatchReceivedRequestAndWaitForReceiptAsync(OnERC1155BatchReceivedFunction onERC1155BatchReceivedFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC1155BatchReceivedFunction, cancellationToken);
        }

        public Task<string> OnERC1155BatchReceivedRequestAsync(string returnValue1, string returnValue2, List<BigInteger> returnValue3, List<BigInteger> returnValue4, byte[] returnValue5)
        {
            var onERC1155BatchReceivedFunction = new OnERC1155BatchReceivedFunction();
                onERC1155BatchReceivedFunction.ReturnValue1 = returnValue1;
                onERC1155BatchReceivedFunction.ReturnValue2 = returnValue2;
                onERC1155BatchReceivedFunction.ReturnValue3 = returnValue3;
                onERC1155BatchReceivedFunction.ReturnValue4 = returnValue4;
                onERC1155BatchReceivedFunction.ReturnValue5 = returnValue5;
            
             return ContractHandler.SendRequestAsync(onERC1155BatchReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC1155BatchReceivedRequestAndWaitForReceiptAsync(string returnValue1, string returnValue2, List<BigInteger> returnValue3, List<BigInteger> returnValue4, byte[] returnValue5, CancellationTokenSource cancellationToken = null)
        {
            var onERC1155BatchReceivedFunction = new OnERC1155BatchReceivedFunction();
                onERC1155BatchReceivedFunction.ReturnValue1 = returnValue1;
                onERC1155BatchReceivedFunction.ReturnValue2 = returnValue2;
                onERC1155BatchReceivedFunction.ReturnValue3 = returnValue3;
                onERC1155BatchReceivedFunction.ReturnValue4 = returnValue4;
                onERC1155BatchReceivedFunction.ReturnValue5 = returnValue5;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC1155BatchReceivedFunction, cancellationToken);
        }

        public Task<string> OnERC1155ReceivedRequestAsync(OnERC1155ReceivedFunction onERC1155ReceivedFunction)
        {
             return ContractHandler.SendRequestAsync(onERC1155ReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC1155ReceivedRequestAndWaitForReceiptAsync(OnERC1155ReceivedFunction onERC1155ReceivedFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC1155ReceivedFunction, cancellationToken);
        }

        public Task<string> OnERC1155ReceivedRequestAsync(string returnValue1, string returnValue2, BigInteger returnValue3, BigInteger returnValue4, byte[] returnValue5)
        {
            var onERC1155ReceivedFunction = new OnERC1155ReceivedFunction();
                onERC1155ReceivedFunction.ReturnValue1 = returnValue1;
                onERC1155ReceivedFunction.ReturnValue2 = returnValue2;
                onERC1155ReceivedFunction.ReturnValue3 = returnValue3;
                onERC1155ReceivedFunction.ReturnValue4 = returnValue4;
                onERC1155ReceivedFunction.ReturnValue5 = returnValue5;
            
             return ContractHandler.SendRequestAsync(onERC1155ReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC1155ReceivedRequestAndWaitForReceiptAsync(string returnValue1, string returnValue2, BigInteger returnValue3, BigInteger returnValue4, byte[] returnValue5, CancellationTokenSource cancellationToken = null)
        {
            var onERC1155ReceivedFunction = new OnERC1155ReceivedFunction();
                onERC1155ReceivedFunction.ReturnValue1 = returnValue1;
                onERC1155ReceivedFunction.ReturnValue2 = returnValue2;
                onERC1155ReceivedFunction.ReturnValue3 = returnValue3;
                onERC1155ReceivedFunction.ReturnValue4 = returnValue4;
                onERC1155ReceivedFunction.ReturnValue5 = returnValue5;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC1155ReceivedFunction, cancellationToken);
        }

        public Task<string> OnERC721ReceivedRequestAsync(OnERC721ReceivedFunction onERC721ReceivedFunction)
        {
             return ContractHandler.SendRequestAsync(onERC721ReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC721ReceivedRequestAndWaitForReceiptAsync(OnERC721ReceivedFunction onERC721ReceivedFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC721ReceivedFunction, cancellationToken);
        }

        public Task<string> OnERC721ReceivedRequestAsync(string returnValue1, string returnValue2, BigInteger returnValue3, byte[] returnValue4)
        {
            var onERC721ReceivedFunction = new OnERC721ReceivedFunction();
                onERC721ReceivedFunction.ReturnValue1 = returnValue1;
                onERC721ReceivedFunction.ReturnValue2 = returnValue2;
                onERC721ReceivedFunction.ReturnValue3 = returnValue3;
                onERC721ReceivedFunction.ReturnValue4 = returnValue4;
            
             return ContractHandler.SendRequestAsync(onERC721ReceivedFunction);
        }

        public Task<TransactionReceipt> OnERC721ReceivedRequestAndWaitForReceiptAsync(string returnValue1, string returnValue2, BigInteger returnValue3, byte[] returnValue4, CancellationTokenSource cancellationToken = null)
        {
            var onERC721ReceivedFunction = new OnERC721ReceivedFunction();
                onERC721ReceivedFunction.ReturnValue1 = returnValue1;
                onERC721ReceivedFunction.ReturnValue2 = returnValue2;
                onERC721ReceivedFunction.ReturnValue3 = returnValue3;
                onERC721ReceivedFunction.ReturnValue4 = returnValue4;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(onERC721ReceivedFunction, cancellationToken);
        }

        public Task<BigInteger> ProposalDeadlineQueryAsync(ProposalDeadlineFunction proposalDeadlineFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ProposalDeadlineFunction, BigInteger>(proposalDeadlineFunction, blockParameter);
        }

        
        public Task<BigInteger> ProposalDeadlineQueryAsync(BigInteger proposalId, BlockParameter blockParameter = null)
        {
            var proposalDeadlineFunction = new ProposalDeadlineFunction();
                proposalDeadlineFunction.ProposalId = proposalId;
            
            return ContractHandler.QueryAsync<ProposalDeadlineFunction, BigInteger>(proposalDeadlineFunction, blockParameter);
        }

        public Task<BigInteger> ProposalEtaQueryAsync(ProposalEtaFunction proposalEtaFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ProposalEtaFunction, BigInteger>(proposalEtaFunction, blockParameter);
        }

        
        public Task<BigInteger> ProposalEtaQueryAsync(BigInteger proposalId, BlockParameter blockParameter = null)
        {
            var proposalEtaFunction = new ProposalEtaFunction();
                proposalEtaFunction.ProposalId = proposalId;
            
            return ContractHandler.QueryAsync<ProposalEtaFunction, BigInteger>(proposalEtaFunction, blockParameter);
        }

        public Task<BigInteger> ProposalSnapshotQueryAsync(ProposalSnapshotFunction proposalSnapshotFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ProposalSnapshotFunction, BigInteger>(proposalSnapshotFunction, blockParameter);
        }

        
        public Task<BigInteger> ProposalSnapshotQueryAsync(BigInteger proposalId, BlockParameter blockParameter = null)
        {
            var proposalSnapshotFunction = new ProposalSnapshotFunction();
                proposalSnapshotFunction.ProposalId = proposalId;
            
            return ContractHandler.QueryAsync<ProposalSnapshotFunction, BigInteger>(proposalSnapshotFunction, blockParameter);
        }

        public Task<BigInteger> ProposalThresholdQueryAsync(ProposalThresholdFunction proposalThresholdFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ProposalThresholdFunction, BigInteger>(proposalThresholdFunction, blockParameter);
        }

        
        public Task<BigInteger> ProposalThresholdQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<ProposalThresholdFunction, BigInteger>(null, blockParameter);
        }

        public Task<ProposalVotesOutputDTO> ProposalVotesQueryAsync(ProposalVotesFunction proposalVotesFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryDeserializingToObjectAsync<ProposalVotesFunction, ProposalVotesOutputDTO>(proposalVotesFunction, blockParameter);
        }

        public Task<ProposalVotesOutputDTO> ProposalVotesQueryAsync(BigInteger proposalId, BlockParameter blockParameter = null)
        {
            var proposalVotesFunction = new ProposalVotesFunction();
                proposalVotesFunction.ProposalId = proposalId;
            
            return ContractHandler.QueryDeserializingToObjectAsync<ProposalVotesFunction, ProposalVotesOutputDTO>(proposalVotesFunction, blockParameter);
        }

        public Task<string> ProposeRequestAsync(ProposeFunction proposeFunction)
        {
             return ContractHandler.SendRequestAsync(proposeFunction);
        }

        public Task<TransactionReceipt> ProposeRequestAndWaitForReceiptAsync(ProposeFunction proposeFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(proposeFunction, cancellationToken);
        }

        public Task<string> ProposeRequestAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, string description)
        {
            var proposeFunction = new ProposeFunction();
                proposeFunction.Targets = targets;
                proposeFunction.Values = values;
                proposeFunction.Calldatas = calldatas;
                proposeFunction.Description = description;
            
             return ContractHandler.SendRequestAsync(proposeFunction);
        }

        public Task<TransactionReceipt> ProposeRequestAndWaitForReceiptAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, string description, CancellationTokenSource cancellationToken = null)
        {
            var proposeFunction = new ProposeFunction();
                proposeFunction.Targets = targets;
                proposeFunction.Values = values;
                proposeFunction.Calldatas = calldatas;
                proposeFunction.Description = description;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(proposeFunction, cancellationToken);
        }

        public Task<string> QueueRequestAsync(QueueFunction queueFunction)
        {
             return ContractHandler.SendRequestAsync(queueFunction);
        }

        public Task<TransactionReceipt> QueueRequestAndWaitForReceiptAsync(QueueFunction queueFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(queueFunction, cancellationToken);
        }

        public Task<string> QueueRequestAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, byte[] descriptionHash)
        {
            var queueFunction = new QueueFunction();
                queueFunction.Targets = targets;
                queueFunction.Values = values;
                queueFunction.Calldatas = calldatas;
                queueFunction.DescriptionHash = descriptionHash;
            
             return ContractHandler.SendRequestAsync(queueFunction);
        }

        public Task<TransactionReceipt> QueueRequestAndWaitForReceiptAsync(List<string> targets, List<BigInteger> values, List<byte[]> calldatas, byte[] descriptionHash, CancellationTokenSource cancellationToken = null)
        {
            var queueFunction = new QueueFunction();
                queueFunction.Targets = targets;
                queueFunction.Values = values;
                queueFunction.Calldatas = calldatas;
                queueFunction.DescriptionHash = descriptionHash;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(queueFunction, cancellationToken);
        }

        public Task<BigInteger> QuorumQueryAsync(QuorumFunction quorumFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<QuorumFunction, BigInteger>(quorumFunction, blockParameter);
        }

        
        public Task<BigInteger> QuorumQueryAsync(BigInteger blockNumber, BlockParameter blockParameter = null)
        {
            var quorumFunction = new QuorumFunction();
                quorumFunction.BlockNumber = blockNumber;
            
            return ContractHandler.QueryAsync<QuorumFunction, BigInteger>(quorumFunction, blockParameter);
        }

        public Task<BigInteger> QuorumDenominatorQueryAsync(QuorumDenominatorFunction quorumDenominatorFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<QuorumDenominatorFunction, BigInteger>(quorumDenominatorFunction, blockParameter);
        }

        
        public Task<BigInteger> QuorumDenominatorQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<QuorumDenominatorFunction, BigInteger>(null, blockParameter);
        }

        public Task<BigInteger> QuorumNumeratorQueryAsync(QuorumNumeratorFunction quorumNumeratorFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<QuorumNumeratorFunction, BigInteger>(quorumNumeratorFunction, blockParameter);
        }

        
        public Task<BigInteger> QuorumNumeratorQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<QuorumNumeratorFunction, BigInteger>(null, blockParameter);
        }

        public Task<string> RelayRequestAsync(RelayFunction relayFunction)
        {
             return ContractHandler.SendRequestAsync(relayFunction);
        }

        public Task<TransactionReceipt> RelayRequestAndWaitForReceiptAsync(RelayFunction relayFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(relayFunction, cancellationToken);
        }

        public Task<string> RelayRequestAsync(string target, BigInteger value, byte[] data)
        {
            var relayFunction = new RelayFunction();
                relayFunction.Target = target;
                relayFunction.Value = value;
                relayFunction.Data = data;
            
             return ContractHandler.SendRequestAsync(relayFunction);
        }

        public Task<TransactionReceipt> RelayRequestAndWaitForReceiptAsync(string target, BigInteger value, byte[] data, CancellationTokenSource cancellationToken = null)
        {
            var relayFunction = new RelayFunction();
                relayFunction.Target = target;
                relayFunction.Value = value;
                relayFunction.Data = data;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(relayFunction, cancellationToken);
        }

        public Task<string> SetProposalThresholdRequestAsync(SetProposalThresholdFunction setProposalThresholdFunction)
        {
             return ContractHandler.SendRequestAsync(setProposalThresholdFunction);
        }

        public Task<TransactionReceipt> SetProposalThresholdRequestAndWaitForReceiptAsync(SetProposalThresholdFunction setProposalThresholdFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setProposalThresholdFunction, cancellationToken);
        }

        public Task<string> SetProposalThresholdRequestAsync(BigInteger newProposalThreshold)
        {
            var setProposalThresholdFunction = new SetProposalThresholdFunction();
                setProposalThresholdFunction.NewProposalThreshold = newProposalThreshold;
            
             return ContractHandler.SendRequestAsync(setProposalThresholdFunction);
        }

        public Task<TransactionReceipt> SetProposalThresholdRequestAndWaitForReceiptAsync(BigInteger newProposalThreshold, CancellationTokenSource cancellationToken = null)
        {
            var setProposalThresholdFunction = new SetProposalThresholdFunction();
                setProposalThresholdFunction.NewProposalThreshold = newProposalThreshold;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setProposalThresholdFunction, cancellationToken);
        }

        public Task<string> SetVotingDelayRequestAsync(SetVotingDelayFunction setVotingDelayFunction)
        {
             return ContractHandler.SendRequestAsync(setVotingDelayFunction);
        }

        public Task<TransactionReceipt> SetVotingDelayRequestAndWaitForReceiptAsync(SetVotingDelayFunction setVotingDelayFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setVotingDelayFunction, cancellationToken);
        }

        public Task<string> SetVotingDelayRequestAsync(BigInteger newVotingDelay)
        {
            var setVotingDelayFunction = new SetVotingDelayFunction();
                setVotingDelayFunction.NewVotingDelay = newVotingDelay;
            
             return ContractHandler.SendRequestAsync(setVotingDelayFunction);
        }

        public Task<TransactionReceipt> SetVotingDelayRequestAndWaitForReceiptAsync(BigInteger newVotingDelay, CancellationTokenSource cancellationToken = null)
        {
            var setVotingDelayFunction = new SetVotingDelayFunction();
                setVotingDelayFunction.NewVotingDelay = newVotingDelay;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setVotingDelayFunction, cancellationToken);
        }

        public Task<string> SetVotingPeriodRequestAsync(SetVotingPeriodFunction setVotingPeriodFunction)
        {
             return ContractHandler.SendRequestAsync(setVotingPeriodFunction);
        }

        public Task<TransactionReceipt> SetVotingPeriodRequestAndWaitForReceiptAsync(SetVotingPeriodFunction setVotingPeriodFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setVotingPeriodFunction, cancellationToken);
        }

        public Task<string> SetVotingPeriodRequestAsync(BigInteger newVotingPeriod)
        {
            var setVotingPeriodFunction = new SetVotingPeriodFunction();
                setVotingPeriodFunction.NewVotingPeriod = newVotingPeriod;
            
             return ContractHandler.SendRequestAsync(setVotingPeriodFunction);
        }

        public Task<TransactionReceipt> SetVotingPeriodRequestAndWaitForReceiptAsync(BigInteger newVotingPeriod, CancellationTokenSource cancellationToken = null)
        {
            var setVotingPeriodFunction = new SetVotingPeriodFunction();
                setVotingPeriodFunction.NewVotingPeriod = newVotingPeriod;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(setVotingPeriodFunction, cancellationToken);
        }

        public Task<byte> StateQueryAsync(StateFunction stateFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<StateFunction, byte>(stateFunction, blockParameter);
        }

        
        public Task<byte> StateQueryAsync(BigInteger proposalId, BlockParameter blockParameter = null)
        {
            var stateFunction = new StateFunction();
                stateFunction.ProposalId = proposalId;
            
            return ContractHandler.QueryAsync<StateFunction, byte>(stateFunction, blockParameter);
        }

        public Task<bool> SupportsInterfaceQueryAsync(SupportsInterfaceFunction supportsInterfaceFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<SupportsInterfaceFunction, bool>(supportsInterfaceFunction, blockParameter);
        }

        
        public Task<bool> SupportsInterfaceQueryAsync(byte[] interfaceId, BlockParameter blockParameter = null)
        {
            var supportsInterfaceFunction = new SupportsInterfaceFunction();
                supportsInterfaceFunction.InterfaceId = interfaceId;
            
            return ContractHandler.QueryAsync<SupportsInterfaceFunction, bool>(supportsInterfaceFunction, blockParameter);
        }

        public Task<string> TimelockQueryAsync(TimelockFunction timelockFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<TimelockFunction, string>(timelockFunction, blockParameter);
        }

        
        public Task<string> TimelockQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<TimelockFunction, string>(null, blockParameter);
        }

        public Task<string> TokenQueryAsync(TokenFunction tokenFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<TokenFunction, string>(tokenFunction, blockParameter);
        }

        
        public Task<string> TokenQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<TokenFunction, string>(null, blockParameter);
        }

        public Task<string> UpdateQuorumNumeratorRequestAsync(UpdateQuorumNumeratorFunction updateQuorumNumeratorFunction)
        {
             return ContractHandler.SendRequestAsync(updateQuorumNumeratorFunction);
        }

        public Task<TransactionReceipt> UpdateQuorumNumeratorRequestAndWaitForReceiptAsync(UpdateQuorumNumeratorFunction updateQuorumNumeratorFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(updateQuorumNumeratorFunction, cancellationToken);
        }

        public Task<string> UpdateQuorumNumeratorRequestAsync(BigInteger newQuorumNumerator)
        {
            var updateQuorumNumeratorFunction = new UpdateQuorumNumeratorFunction();
                updateQuorumNumeratorFunction.NewQuorumNumerator = newQuorumNumerator;
            
             return ContractHandler.SendRequestAsync(updateQuorumNumeratorFunction);
        }

        public Task<TransactionReceipt> UpdateQuorumNumeratorRequestAndWaitForReceiptAsync(BigInteger newQuorumNumerator, CancellationTokenSource cancellationToken = null)
        {
            var updateQuorumNumeratorFunction = new UpdateQuorumNumeratorFunction();
                updateQuorumNumeratorFunction.NewQuorumNumerator = newQuorumNumerator;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(updateQuorumNumeratorFunction, cancellationToken);
        }

        public Task<string> UpdateTimelockRequestAsync(UpdateTimelockFunction updateTimelockFunction)
        {
             return ContractHandler.SendRequestAsync(updateTimelockFunction);
        }

        public Task<TransactionReceipt> UpdateTimelockRequestAndWaitForReceiptAsync(UpdateTimelockFunction updateTimelockFunction, CancellationTokenSource cancellationToken = null)
        {
             return ContractHandler.SendRequestAndWaitForReceiptAsync(updateTimelockFunction, cancellationToken);
        }

        public Task<string> UpdateTimelockRequestAsync(string newTimelock)
        {
            var updateTimelockFunction = new UpdateTimelockFunction();
                updateTimelockFunction.NewTimelock = newTimelock;
            
             return ContractHandler.SendRequestAsync(updateTimelockFunction);
        }

        public Task<TransactionReceipt> UpdateTimelockRequestAndWaitForReceiptAsync(string newTimelock, CancellationTokenSource cancellationToken = null)
        {
            var updateTimelockFunction = new UpdateTimelockFunction();
                updateTimelockFunction.NewTimelock = newTimelock;
            
             return ContractHandler.SendRequestAndWaitForReceiptAsync(updateTimelockFunction, cancellationToken);
        }

        public Task<string> VersionQueryAsync(VersionFunction versionFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VersionFunction, string>(versionFunction, blockParameter);
        }

        
        public Task<string> VersionQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VersionFunction, string>(null, blockParameter);
        }

        public Task<BigInteger> VotingDelayQueryAsync(VotingDelayFunction votingDelayFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VotingDelayFunction, BigInteger>(votingDelayFunction, blockParameter);
        }

        
        public Task<BigInteger> VotingDelayQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VotingDelayFunction, BigInteger>(null, blockParameter);
        }

        public Task<BigInteger> VotingPeriodQueryAsync(VotingPeriodFunction votingPeriodFunction, BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VotingPeriodFunction, BigInteger>(votingPeriodFunction, blockParameter);
        }

        
        public Task<BigInteger> VotingPeriodQueryAsync(BlockParameter blockParameter = null)
        {
            return ContractHandler.QueryAsync<VotingPeriodFunction, BigInteger>(null, blockParameter);
        }
    }
}
