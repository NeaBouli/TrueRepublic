using System;
using System.IO;

namespace Pnyx.SmartContracts.Solidity.Contracts
{
    public static class BytecodeProvider
    {
        private const string FILE_EXTENSION_BINARY = "bin";

        public static string Load(string contractGroup, string contractName)
        {
            string contractBinaryRoot = Pnyx_SmartContracts_Solidity.Default.ContractBinaryRoot;
            string binaryPath = String.Format("{0}.{1}", Path.Combine(contractBinaryRoot, contractGroup, contractName), FILE_EXTENSION_BINARY);
            if (File.Exists(binaryPath))
            {
                return File.ReadAllText(binaryPath);
            }
            throw new FileNotFoundException(String.Format("Contract file \"{0}\" not found", binaryPath));
        }
    }
}
