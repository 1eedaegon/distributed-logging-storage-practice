package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"google.golang.org/protobuf/proto"
)

// Log는 segment단위로 관리한다.
// Segment는 record를 추가할 때
// 데이터를 store에 쓰고 index를 추가한다.
// config을 참고해서 store/index 파일이 가득찼는 지 확인한다.
type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, // ReadWrite면서 없으면 생성, Memory에서 이어쓰기
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE, // ReadWrite면서 없으면 생성
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset // index가 비어있으면 baseOffset으로 설정
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1 // index가 비어있지 않으면 baseoOffset + 상대 offset + 1
	}
	return s, nil

}

func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}
	if err = s.index.Write(
		uint32(s.nextOffset-uint64(s.baseOffset)), // 상대 offset
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segment) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(off - s.baseOffset)) // 상대 offset으로 index position을 가져온다.
	if err != nil {
		return nil, err
	}
	p, err := s.store.Read(pos) // position 값으로 store에서 읽어온다.
	if err != nil {
		return nil, err
	}
	record := &api.Record{}
	err = proto.Unmarshal(p, record) // 읽어온 데이터를 record로 unmarshal 후 return
	return record, err
}

// Index 혹은 Store가 꽉찼는지 확인
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}
