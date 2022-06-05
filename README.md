# go-jellyfish-merkletree
diem jellyfish merkletree的go实现，数据存储使用leveldb，可用于区块链状态数据的记录。diem项目：[diem](https://github.com/diem/diem)
，参考论文：[jellyfish-merkle-tree](https://diem-developers-components.netlify.app/papers/jellyfish-merkle-tree/2021-01-14.pdf)

对外封装了以下接口，可直接调用：
* [PutValueSet](https://github.com/rjkris/go-jellyfish-merkletree/blob/main/jellyfish/jellyfish_merkletree.go) :更新单个账户数据
* [PutValueSets](https://github.com/rjkris/go-jellyfish-merkletree/blob/main/jellyfish/jellyfish_merkletree.go) :批量更新账户数据
* [GetWithProof](https://github.com/rjkris/go-jellyfish-merkletree/blob/main/jellyfish/jellyfish_merkletree.go) :获取有效性证明
* [Verify](https://github.com/rjkris/go-jellyfish-merkletree/blob/main/jellyfish/proof.go) :有效性验证
