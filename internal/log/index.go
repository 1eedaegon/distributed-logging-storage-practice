package log

import (
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4                   // 레코드 크기
	posWidth uint64 = 8                   // 위치
	entWidth        = offWidth + posWidth // 오프셋 위치
)

type index struct {
	file *os.File
	mmap gommap.MMap // Golang은 메모리 맵 구현이 까다롭다. 물론 다른 언어도 마찬가지겠지만
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	file, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(file.Size())
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxStoreBytes)); err != nil {
		return nil, err
	}
	// #include <sys/mman.h>와 동작이 같다.
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return idx, nil
}
