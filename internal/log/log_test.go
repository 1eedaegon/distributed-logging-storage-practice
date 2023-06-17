package log

import (
	"io/ioutil"
	"os"
	"testing"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	// Table testing: 각 테스트 케이스를 테이블로 만들어서 테스트
	for scenario, fn := range map[string]func(t *testing.T, log *Log){
		"append and read a record succeeds": testAppendRead,
		"offset out of range error":         testOutOfRange,
		"init with existing segments":       testInitExisting,
		"reader":                            testReader,
		"truncate":                          testTruncate,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := Config{}
			c.Segment.MaxStoreBytes = 32
			if scenario == "make new segment" {
				c.Segment.MaxIndexBytes = 13
			}
			log, err := NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	oneRecord := &api.Record{Value: []byte("Hello world")}
	off, err := log.Append(oneRecord)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)
	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, oneRecord.Value, read.Value)
}

func testOutOfRange(t *testing.T, log *Log) {
	segment, err := log.Read(1)
	require.Nil(t, segment) // 범위가 넘어가면 nil을 반환해야 한다.
	require.Error(t, err)   // Out of range가 나와야 한다.
}

func testInitExisting(t *testing.T, log *Log) {
	appendRecord := &api.Record{Value: []byte("Hello world")}
	for i := 0; i < 3; i++ {
		_, err := log.Append(appendRecord)
		require.NoError(t, err)
	}
	require.NoError(t, log.Close())
	offset, err := log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	offset, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)

	// 이전에 만든 log를 다시 불러온다.
	l, err := NewLog(log.Dir, log.Config)
	require.NoError(t, err)

	offset, err = l.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	offset, err = l.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)
}

func testReader(t *testing.T, log *Log) {
	appendRecord := &api.Record{Value: []byte("Hello world!")}
	offset, err := log.Append(appendRecord)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	reader := log.Reader()
	b, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, appendRecord.Value, read.Value)
}
func testTruncate(t *testing.T, log *Log) {
	appendRecord := &api.Record{Value: []byte("Log testing!")}
	for i := 0; i < 3; i++ {
		_, err := log.Append(appendRecord)
		require.NoError(t, err)
	}

	_, err := log.Read(0)
	require.NoError(t, err)

	err = log.Truncate(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err) // offset이 0인 record는 이미 지워졌으므로 에러가 나와야 한다.
}
