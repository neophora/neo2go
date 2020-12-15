package network

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/neophora/neo2go/pkg/core"
	"github.com/neophora/neo2go/pkg/core/block"
	"go.uber.org/zap"
)

type blockQueue struct {
	log         *zap.Logger
	queue       *queue.PriorityQueue
	checkBlocks chan struct{}
	chain       core.Blockchainer
	relayF      func(*block.Block)
}

func newBlockQueue(capacity int, bc core.Blockchainer, log *zap.Logger, relayer func(*block.Block)) *blockQueue {
	if log == nil {
		return nil
	}

	return &blockQueue{
		log:         log,
		queue:       queue.NewPriorityQueue(capacity, false),
		checkBlocks: make(chan struct{}, 1),
		chain:       bc,
		relayF:      relayer,
	}
}

func (bq *blockQueue) run() {
	for {
		_, ok := <-bq.checkBlocks
		if !ok {
			break
		}
		for {
			item := bq.queue.Peek()
			if item == nil {
				break
			}
			minblock := item.(*block.Block)
			if minblock.Index <= bq.chain.BlockHeight()+1 {
				_, _ = bq.queue.Get(1)
				updateBlockQueueLenMetric(bq.length())
				if minblock.Index == bq.chain.BlockHeight()+1 {
					err := bq.chain.AddBlock(minblock)
					if err != nil {
						// The block might already be added by consensus.
						if _, errget := bq.chain.GetBlock(minblock.Hash()); errget != nil {
							bq.log.Warn("blockQueue: failed adding block into the blockchain",
								zap.String("error", err.Error()),
								zap.Uint32("blockHeight", bq.chain.BlockHeight()),
								zap.Uint32("nextIndex", minblock.Index))
						}
					} else if bq.relayF != nil {
						bq.relayF(minblock)
					}
				}
			} else {
				break
			}
		}
	}
}

func (bq *blockQueue) putBlock(block *block.Block) error {
	if bq.chain.BlockHeight() >= block.Index {
		// can easily happen when fetching the same blocks from
		// different peers, thus not considered as error
		return nil
	}
	err := bq.queue.Put(block)
	// update metrics
	updateBlockQueueLenMetric(bq.length())
	select {
	case bq.checkBlocks <- struct{}{}:
		// ok, signalled to goroutine processing queue
	default:
		// it's already busy processing blocks
	}
	return err
}

func (bq *blockQueue) discard() {
	close(bq.checkBlocks)
	bq.queue.Dispose()
}

func (bq *blockQueue) length() int {
	return bq.queue.Len()
}
