package log

/*
	- 로그 -
	- Record: 로그에 저장된 데이터
	- Store file: 레코드를 저장한 파일
	- Index file: 저장 파일을 Index한 파일
	- Segment: Index + Store file
	- Log: 세그먼트 집합
*/

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// 버퍼를 위한 바이너리 순서를 빅앤디안으로
var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	file, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(file.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	// 버퍼에 레코드 길이랑 같이 쓰기
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	// 버퍼 쓰기, 시스템 콜 부하 감소
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

// pos부터
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 남아있는 버퍼 flush
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// offset부터 len(p)만큼 Read
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// 파일 닫기 전 flush
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
