package orderer

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric/common/ledger/blkstorage"
	"github.com/hyperledger/fabric/common/ledger/blkstorage/fsblkstorage"
	"github.com/hyperledger/fabric/common/localmsp"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/ledger/ledgerconfig"
	"github.com/hyperledger/fabric/msp/mgmt"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/utils"
	"path"
)

func BlockStoreProdiver(ordererPath, mspPath string) blkstorage.BlockStoreProvider {
	x := factory.GetDefaultOpts()
	if mspPath != "" {
		x.SwOpts.Ephemeral = false
		x.SwOpts.FileKeystore = &factory.FileKeystoreOpts{
			KeyStorePath: path.Join(mspPath, "keystore"),
			//KeyStorePath: "/Users/d070098/Projects/scp-bc-fabric-node-manager/test/fixtures/node1/crypto/ordererOrganizations/sap.com/orderers/orderer0.sap.com/msp/keystore",
		}
	}
	if err := factory.InitFactories(x); err != nil {
		panic(err)
	}

	attrsToIndex := []blkstorage.IndexableAttr{
		//blkstorage.IndexableAttrBlockHash,
		blkstorage.IndexableAttrBlockNum,
		//blkstorage.IndexableAttrTxID,
		//blkstorage.IndexableAttrBlockNumTranNum,
		//blkstorage.IndexableAttrBlockTxID,
		//blkstorage.IndexableAttrTxValidationCode,
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

func CreateNoOpBlock(lastBlock *common.Block, kafkaMetadata *orderer.KafkaMetadata, mspID, mspPath string) (*common.Block, error) {
	//flogging.SetModuleLevel("msp", "DEBUG")
	if err := mgmt.LoadLocalMsp(mspPath, nil, mspID); err != nil {
		return nil, fmt.Errorf("failed to LoadLocalMsp: %v", err)
	}
	envelope := &common.Envelope{}
	if err := proto.Unmarshal(lastBlock.Data.Data[0], envelope); err != nil {
		return nil, fmt.Errorf("error reconstructing envelope(%s)", err)
	}
	chdr, err := utils.ChannelHeader(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to get lastBlock's ChannelHeader: %v", err)
	}

	block := common.NewBlock(lastBlock.Header.Number+1, lastBlock.Header.DataHash)
	//for _, tx := range transactions {
	//	txEnvBytes, err := proto.Marshal(tx)
	//	if err != nil {
	//		return nil, err
	//	}
	//	block.Data.Data = append(block.Data.Data, txEnvBytes)
	//}
	block.Data.Data = append(block.Data.Data,
		utils.MarshalOrPanic(&common.Envelope{
			Payload: utils.MarshalOrPanic(&common.Payload{
				Header: &common.Header{
					ChannelHeader: utils.MarshalOrPanic(&common.ChannelHeader{
						Type:      int32(common.HeaderType_MESSAGE),
						ChannelId: chdr.ChannelId,
					}),
				},
			}),
		}))
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

	signer := localmsp.NewSigner()
	addBlockSignature(signer, block)

	return block, nil
}

func addBlockSignature(signer crypto.LocalSigner, block *common.Block) {
	blockSignature := &common.MetadataSignature{
		SignatureHeader: utils.MarshalOrPanic(utils.NewSignatureHeaderOrPanic(signer)),
	}

	// Note, this value is intentionally nil, as this metadata is only about the signature, there is no additional metadata
	// information required beyond the fact that the metadata item is signed.
	blockSignatureValue := []byte(nil)

	blockSignature.Signature = utils.SignOrPanic(signer, util.ConcatenateBytes(blockSignatureValue, blockSignature.SignatureHeader, block.Header.Bytes()))

	block.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES] = utils.MarshalOrPanic(&common.Metadata{
		Value: blockSignatureValue,
		Signatures: []*common.MetadataSignature{
			blockSignature,
		},
	})
}

func AddBlock(blockStoreProvider blkstorage.BlockStoreProvider, ledger string, block *common.Block) error {
	blockStore, err := blockStoreProvider.OpenBlockStore(ledger)
	if err != nil {
		return fmt.Errorf("failed to OpenBlockStore: %v", err)
	}
	return blockStore.AddBlock(block)
}
