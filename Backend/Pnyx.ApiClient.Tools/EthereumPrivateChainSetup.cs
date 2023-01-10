namespace Pnyx.ApiClient.Tools
{
    public class EthereumPrivateChainSetup
    {
        private const string GENESIS_FILE_NAME = "genesis.json";

        private static string[] GENESIS_JSON_TEMPLATE = {
              "{\r\n"
                  , "\"config\": {\r\n"
                    , "\"chainId\": {0},\r\n"
                    , "\"homesteadBlock\": 0,\r\n"
                    , "\"eip150Block\": 0,\r\n"
                    , "\"eip150Hash\": \"0x0000000000000000000000000000000000000000000000000000000000000000\",\r\n"
                    , "\"eip155Block\": 0,\r\n"
                    , "\"eip158Block\": 0,\r\n"
                    , "\"byzantiumBlock\": 0,\r\n"
                    , "\"constantinopleBlock\": 0,\r\n"
                    , "\"petersburgBlock\": 0,\r\n"
                    , "\"istanbulBlock\": 0,\r\n"
                    , "\"ethash\": { }\r\n"
                  , "}\r\n"
                , "\"nonce\": \"0x0\"\r\n"
                , "\"timestamp\": \"0x0\"\r\n"
                , "\"extraData\": \"0x0000000000000000000000000000000000000000000000000000000000000000\"\r\n"
                , "\"gasLimit\": \"0x0\"\r\n"
                , "\"difficulty\": \"0x0\"\r\n"
                , "\"mixHash\": \"0x0000000000000000000000000000000000000000000000000000000000000000\"\r\n"
                , "\"coinbase\": \"0x0000000000000000000000000000000000000000\"\r\n"
                , "\"alloc\": {\r\n"
                        , "\"0000000000000000000000000000000000000088\": {\r\n"
                        , "\"balance\": \"0x200000000000000000000000000000000000000000000000000000000000000\"\r\n"
                    , "}\r\n"
                , "}\r\n"
                , "\"number\": \"0x0\"\r\n"
                , "\"gasUsed\": \"0x0\"\r\n"
                , "\"parentHash\": \"0x0000000000000000000000000000000000000000000000000000000000000000\"\r\n"
            , "}\r\n"
        };

        public static void CreateGenesisJson(string path, int chainId, string difficulty)
        {
            if (!Directory.Exists(path))
            {

            }
        }
    }
}