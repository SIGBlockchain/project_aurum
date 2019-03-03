package blockchain

import block "../block"

func AddBlock(b block.Block, filename string) error {

	return nil
}

func GetBlockByHeight(height int, filename string) (block.Block, error) {

	return block.Block{}, nil
}

func GetBlockByIndex(index, filename string) (block.Block, error) {

	return block.Block{}, nil
}

func GetBlockByHash(hash []byte, filename string) (block.Block, error) {

	return block.Block{}, nil
}
