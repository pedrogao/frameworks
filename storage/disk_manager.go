package storage

import (
	"fmt"
	"os"
)

type pgnum uint64

type Options struct {
	pageSize int

	MinFillPercent float32
	MaxFillPercent float32
}

var DefaultOptions = &Options{
	MinFillPercent: 0.5,
	MaxFillPercent: 0.95,
}

// page of disk
type page struct {
	num  pgnum
	data []byte
}

type diskManager struct {
	pageSize       int
	minFillPercent float32
	maxFillPercent float32
	file           *os.File

	*meta     // meta data
	*freelist // free list
}

func newDiskManager(path string, options *Options) (*diskManager, error) {
	dm := &diskManager{
		meta:           newEmptyMeta(),
		pageSize:       options.pageSize,
		minFillPercent: options.MinFillPercent,
		maxFillPercent: options.MaxFillPercent,
	}

	// exist
	_, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		// init freelist
		dm.file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			_ = dm.close()
			return nil, err
		}

		dm.freelist = newFreelist()        // create freelist
		dm.freelistPage = dm.getNextPage() // allocate freelist page
		_, err := dm.writeFreelist()       // flush disk
		if err != nil {
			_ = dm.close()
			return nil, err
		}

		// init root node
		collectionsNode, err := dm.writeNode(NewNodeForSerialization([]*Item{}, []pgnum{}))
		if err != nil {
			_ = dm.close()
			return nil, err
		}
		dm.root = collectionsNode.pageNum // set root node page
		_, err = dm.writeMeta(dm.meta)    // write meta page
		if err != nil {
			_ = dm.close()
			return nil, err
		}
	case err != nil && !os.IsNotExist(err):
		return nil, err
	default:
		dm.file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			_ = dm.close()
			return nil, err
		}

		meta, err := dm.readMeta()
		if err != nil {
			_ = dm.close()
			return nil, err
		}
		dm.meta = meta

		freelist, err := dm.readFreelist()
		if err != nil {
			_ = dm.close()
			return nil, err
		}
		dm.freelist = freelist
	}

	return dm, nil
}

// getSplitIndex should be called when performing rebalance after an item is removed. It checks if a node can spare an
// element, and if it does then it returns the index when there the split should happen. Otherwise -1 is returned.
func (d *diskManager) getSplitIndex(node *Node) int {
	size := 0
	size += nodeHeaderSize

	for i := range node.items {
		size += node.elementSize(i)

		// if we have a big enough page size (more than minimum), and didn't reach the last node, which means we can
		// spare an element
		if float32(size) > d.minThreshold() && i < len(node.items)-1 {
			return i + 1
		}
	}

	return -1
}

func (d *diskManager) maxThreshold() float32 {
	return d.maxFillPercent * float32(d.pageSize)
}

func (d *diskManager) isOverPopulated(node *Node) bool {
	return float32(node.nodeSize()) > d.maxThreshold()
}

func (d *diskManager) minThreshold() float32 {
	return d.minFillPercent * float32(d.pageSize)
}

func (d *diskManager) isUnderPopulated(node *Node) bool {
	return float32(node.nodeSize()) < d.minThreshold()
}

func (d *diskManager) close() error {
	if d.file != nil {
		err := d.file.Close()
		if err != nil {
			return fmt.Errorf("could not close file: %s", err)
		}
		d.file = nil
	}

	return nil
}

func (d *diskManager) allocateEmptyPage() *page {
	return &page{
		data: make([]byte, d.pageSize),
	}
}

func (d *diskManager) readPage(pageNum pgnum) (*page, error) {
	p := d.allocateEmptyPage()

	offset := int(pageNum) * d.pageSize
	_, err := d.file.ReadAt(p.data, int64(offset))
	if err != nil {
		return nil, err
	}
	return p, err
}

func (d *diskManager) writePage(p *page) error {
	offset := int64(p.num) * int64(d.pageSize)
	_, err := d.file.WriteAt(p.data, offset)
	return err
}

func (d *diskManager) getNode(pageNum pgnum) (*Node, error) {
	p, err := d.readPage(pageNum)
	if err != nil {
		return nil, err
	}
	node := NewEmptyNode()
	node.deserialize(p.data)
	node.pageNum = pageNum
	return node, nil
}

func (d *diskManager) writeNode(n *Node) (*Node, error) {
	p := d.allocateEmptyPage()
	if n.pageNum == 0 {
		p.num = d.getNextPage()
		n.pageNum = p.num
	} else {
		p.num = n.pageNum
	}

	p.data = n.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (d *diskManager) deleteNode(pageNum pgnum) {
	d.releasePage(pageNum)
}

func (d *diskManager) readFreelist() (*freelist, error) {
	p, err := d.readPage(d.freelistPage)
	if err != nil {
		return nil, err
	}

	freelist := newFreelist()
	freelist.deserialize(p.data)
	return freelist, nil
}

func (d *diskManager) writeFreelist() (*page, error) {
	p := d.allocateEmptyPage()
	p.num = d.freelistPage
	d.freelist.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	d.freelistPage = p.num
	return p, nil
}

func (d *diskManager) writeMeta(meta *meta) (*page, error) {
	p := d.allocateEmptyPage()
	p.num = metaPageNum
	meta.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *diskManager) readMeta() (*meta, error) {
	p, err := d.readPage(metaPageNum)
	if err != nil {
		return nil, err
	}

	meta := newEmptyMeta()
	meta.deserialize(p.data)
	return meta, nil
}
