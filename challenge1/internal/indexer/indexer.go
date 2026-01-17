package indexer

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zacksfF/sepolia-sh/ch1/internal/eth"
	"github.com/zacksfF/sepolia-sh/ch1/internal/model"
	"github.com/zacksfF/sepolia-sh/ch1/internal/storage"
)

type Indexer struct {
	eth      *eth.Client
	store    storage.Store
	contract common.Address
	topic    common.Hash
}

func New(
	ethClient *eth.Client,
	store storage.Store,
	contract common.Address,
	topic common.Hash,
) *Indexer {
	return &Indexer{
		eth:      ethClient,
		store:    store,
		contract: contract,
		topic:    topic,
	}
}

func (i *Indexer) Run(
	ctx context.Context,
	start uint64,
	end *uint64,
	batch uint64,
) error {

	logFetcher := eth.NewLogFetcher(i.eth)
	blockFetcher := eth.NewBlockFetcher(i.eth)

	nextIndex, err := i.store.GetNextIndex(ctx)
	if err != nil {
		return err
	}

	var latest uint64
	if end == nil {
		latest, err = i.eth.BlockNumber(ctx)
		if err != nil {
			return err
		}
	} else {
		latest = *end
	}

	blockCache := make(map[common.Hash]*types.Block)

	for from := start; from <= latest; from += batch {
		to := from + batch - 1
		if to > latest {
			to = latest
		}

		log.Printf("fetching logs: blocks %d -> %d", from, to)

		logs, err := logFetcher.Fetch(ctx, i.contract, i.topic, from, to)
		if err != nil {
			return err
		}

		for _, lg := range logs {
			blk, ok := blockCache[lg.BlockHash]
			if !ok {
				blk, err = blockFetcher.ByHash(ctx, lg.BlockHash)
				if err != nil {
					return err
				}
				blockCache[lg.BlockHash] = blk
			}

			event := &model.IndexedEvent{
				Index:       nextIndex,
				BlockNumber: blk.NumberU64(),
				BlockTime:   blk.Time(),
				ParentHash:  blk.ParentHash(),
				TxHash:      lg.TxHash,
				LogIndex:    lg.Index,
				InfoRoot:    common.BytesToHash(lg.Data), // assumption documented in README
			}

			if err := i.store.SaveEvent(ctx, event); err != nil {
				return err
			}

			nextIndex++
			if err := i.store.SetNextIndex(ctx, nextIndex); err != nil {
				return err
			}
		}
	}

	return nil
}
