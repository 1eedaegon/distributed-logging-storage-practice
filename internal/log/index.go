package log

import (
	"io"
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
	idx.size = uint64(file.Size())                                               // 인덱스에 파일에서 읽은 사이즈 추가
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil { // 인덱스 파일 최대크기로 변경
		return nil, err
	}
	// #include <sys/mman.h>와 동작이 같다.
	// 인덱스에 메모리 맵파일 생성
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return idx, nil
}

// Graceful shutdown
// file과 mmap을 sync시켜서 flush한다.
func (idx *index) Close() error {
	if err := idx.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := idx.file.Sync(); err != nil {
		return err
	}
	if err := idx.file.Truncate(int64(idx.size)); err != nil {
		return err
	}
	return idx.file.Close()
}

// TODO: Ungraceful shutdown - sanity check
// 서비스 시작 시 Sanity check을 해서 손상 데이터를 복구한다.

// Read
// 상대적인 offset을 사용한다.
func (idx *index) Read(in int64) (out uint32, pos uint64, err error) {
	if idx.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((idx.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if idx.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(idx.mmap[pos : pos+offWidth])
	pos = enc.Uint64(idx.mmap[pos+offWidth : idx.size+entWidth])
	return out, pos, nil
}

func (idx *index) Write(off uint32, pos uint64) error {
	if uint64(len(idx.mmap)) < idx.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(idx.mmap[idx.size:idx.size+offWidth], off)
	enc.PutUint64(idx.mmap[idx.size+offWidth:idx.size+entWidth], pos)
	idx.size += uint64(entWidth)
	return nil
}

func (idx *index) Name() string {
	return idx.file.Name()
}
