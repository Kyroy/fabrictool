package orderer

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/ledger/blkstorage"
	"github.com/hyperledger/fabric/common/ledger/blkstorage/fsblkstorage"
	"github.com/hyperledger/fabric/core/ledger/ledgerconfig"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/utils"
)

func BlockStoreProdiver(ordererPath string) blkstorage.BlockStoreProvider {
	factory.InitFactories(nil)

	attrsToIndex := []blkstorage.IndexableAttr{
		blkstorage.IndexableAttrBlockHash,
		blkstorage.IndexableAttrBlockNum,
		blkstorage.IndexableAttrTxID,
		blkstorage.IndexableAttrBlockNumTranNum,
		blkstorage.IndexableAttrBlockTxID,
		blkstorage.IndexableAttrTxValidationCode,
	}
	indexConfig := &blkstorage.IndexConfig{AttrsToIndex: attrsToIndex}
	blockStoreProvider := fsblkstorage.NewProvider(
		fsblkstorage.NewConf(ordererPath, ledgerconfig.GetMaxBlockfileSize()),
		indexConfig)
	//defer blockStoreProvider.Close()
	return blockStoreProvider
}

func ListLedgers(blockStoreProvider blkstorage.BlockStoreProvider) {
	ledgers, err := blockStoreProvider.List()
	if err != nil {
		fmt.Println("err1", err)
		return
	}

	for _, ledger := range ledgers {
		if ledger == "index" {
			continue
		}
		block, err := LastBlock(blockStoreProvider, ledger)
		if err != nil {
			fmt.Println(ledger, err)
			continue
		}
		kafkaMetadata, err := LedgerKafkaMetadata(block, ledger)
		if err != nil {
			fmt.Println(ledger, err)
			continue
		}
		fmt.Printf("%s: last_block %d, last_offset_persisted %d\n", ledger, block.Header.Number, kafkaMetadata.LastOffsetPersisted)
	}
}

func LedgerKafkaMetadata(block *common.Block, ledger string) (*orderer.KafkaMetadata, error) {
	metadata, err := utils.GetMetadataFromBlock(block, common.BlockMetadataIndex_ORDERER)
	if err != nil {
		return nil, fmt.Errorf("failed to GetMetadataFromBlock: %v", err)
	}
	var kafkaMetadata orderer.KafkaMetadata
	if err := proto.Unmarshal(metadata.Value, &kafkaMetadata); err != nil {
		return nil, err
	}
	return &kafkaMetadata, nil
}

func LastBlock(blockStoreProvider blkstorage.BlockStoreProvider, ledger string) (*common.Block, error) {
	blockStore, err := blockStoreProvider.OpenBlockStore(ledger)
	if err != nil {
		return nil, fmt.Errorf("failed to OpenBlockStore: %v", err)
	}
	blockInfo, err := blockStore.GetBlockchainInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to GetBlockchainInfo: %v", err)
	}
	block, err := blockStore.RetrieveBlockByNumber(blockInfo.Height - 1)
	if err != nil {
		return nil, fmt.Errorf("failed to RetrieveBlockByNumber: %v", err)
	}
	return block, nil
}

func CreateNoOpBlock(lastBlock *common.Block, kafkaMetadata *orderer.KafkaMetadata) (*common.Block, error) {
	block := common.NewBlock(lastBlock.Header.Number+1, lastBlock.Header.DataHash)
	//for _, tx := range transactions {
	//	txEnvBytes, _ := proto.Marshal(tx)
	//	block.Data.Data = append(block.Data.Data, txEnvBytes)
	//}
	block.Header.DataHash = block.Data.Hash()
	utils.CopyBlockMetadata(lastBlock, block)

	// insert new KafkaMetadata
	value, err := proto.Marshal(kafkaMetadata)
	if err != nil {
		return nil, err
	}
	md := &common.Metadata{
		Value: value,
	}
	metadata, err := proto.Marshal(md)
	if err != nil {
		return nil, err
	}
	block.Metadata.Metadata[common.BlockMetadataIndex_ORDERER] = metadata

	return block, nil
}

func AddBlock(blockStoreProvider blkstorage.BlockStoreProvider, ledger string, block *common.Block) error {
	blockStore, err := blockStoreProvider.OpenBlockStore(ledger)
	if err != nil {
		return fmt.Errorf("failed to OpenBlockStore: %v", err)
	}
	return blockStore.AddBlock(block)
}
